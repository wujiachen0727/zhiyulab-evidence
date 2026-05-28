#!/usr/bin/env python3
"""汇总 slice 扩容实验输出，生成正文可引用的表格。"""

from __future__ import annotations

from pathlib import Path
import re
from statistics import mean

BASE = Path(__file__).resolve().parents[2]
OUT = BASE / "output" / "slice-growth-compare"
DATA = BASE / "data"


def parse_scan(path: Path):
    rows = []
    for line in path.read_text().splitlines():
        if not line or line.startswith("#") or line.startswith("oldcap"):
            continue
        oldcap, newcap, delta, bytes_, status = line.split(",")
        rows.append({
            "oldcap": int(oldcap),
            "newcap": int(newcap),
            "delta": int(delta),
            "bytes": int(bytes_),
            "status": status,
        })
    return rows


def parse_bench(path: Path):
    rows = {}
    pattern = re.compile(r"^(Benchmark\S+)\s+\d+\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op")
    for line in path.read_text().splitlines():
        m = pattern.match(line.strip())
        if not m:
            continue
        name, ns, b, allocs = m.groups()
        name = name.split("-")[0]
        rows.setdefault(name, {"ns/op": [], "B/op": [], "allocs/op": []})
        rows[name]["ns/op"].append(int(ns))
        rows[name]["B/op"].append(int(b))
        rows[name]["allocs/op"].append(int(allocs))
    return {
        name: {metric: mean(vals) for metric, vals in metrics.items()}
        for name, metrics in rows.items()
    }


def main() -> None:
    go117_scan = parse_scan(OUT / "go117-scan-900-1400.csv")
    go126_scan = parse_scan(OUT / "go126-scan-900-1400.csv")

    go117_downs = [r for r in go117_scan if r["status"] == "down"]
    go126_downs = [r for r in go126_scan if r["status"] == "down"]

    sample_caps = [900, 960, 1000, 1023, 1024, 1025, 1100, 1200, 1300, 1400]
    by_old_117 = {r["oldcap"]: r for r in go117_scan}
    by_old_126 = {r["oldcap"]: r for r in go126_scan}

    bench117 = parse_bench(OUT / "go117-bench.txt")
    bench126 = parse_bench(OUT / "go126-bench.txt")

    lines = []
    lines.append("# Slice 扩容实验摘要")
    lines.append("")
    lines.append("> 数据标注：除特别说明外，以下均为 `[实测 Go 1.17.13 / Go 1.26.2 darwin/arm64]`。")
    lines.append("")
    lines.append("## E1：Go 1.17 非单调增长复现")
    lines.append("")
    lines.append(f"- Go 1.17 扫描 oldcap=900..1400，append 1 个 byte 后，发现下降点 {len(go117_downs)} 个。")
    for r in go117_downs:
        prev = next(x for x in go117_scan if x["oldcap"] == r["oldcap"] - 1)
        lines.append(f"- 关键下降：oldcap {prev['oldcap']} → newcap {prev['newcap']}；oldcap {r['oldcap']} → newcap {r['newcap']}。也就是 oldcap 增加 1，结果 newcap 从 {prev['newcap']} 掉到 {r['newcap']}。")
    lines.append(f"- Go 1.26.2 同范围下降点 {len(go126_downs)} 个。")
    lines.append("")
    lines.append("### 1024 附近窗口")
    lines.append("")
    lines.append("| oldcap | Go 1.17 newcap | Go 1.26 newcap | 说明 |")
    lines.append("|---:|---:|---:|---|")
    for old in range(1018, 1031):
        r117 = by_old_117[old]
        r126 = by_old_126[old]
        note = ""
        if old == 1023:
            note = "旧策略 1024 以下仍翻倍到 2048"
        elif old == 1024:
            note = "旧策略跨过阈值后掉到 1280；新策略保持 1536"
        lines.append(f"| {old} | {r117['newcap']} | {r126['newcap']} | {note} |")

    lines.append("")
    lines.append("## E2：新旧容量序列对比")
    lines.append("")
    lines.append("| oldcap | Go 1.17 append 后 cap | Go 1.26 append 后 cap | 差异 |")
    lines.append("|---:|---:|---:|---:|")
    for old in sample_caps:
        r117 = by_old_117[old]
        r126 = by_old_126[old]
        lines.append(f"| {old} | {r117['newcap']} | {r126['newcap']} | {r126['newcap'] - r117['newcap']} |")

    lines.append("")
    lines.append("## E3：benchmark 量化")
    lines.append("")
    lines.append("> 解读规则：先看 allocs/op 与 B/op，再看 ns/op。")
    lines.append("")
    lines.append("| benchmark | Go 1.17 ns/op | Go 1.26 ns/op | Go 1.17 B/op | Go 1.26 B/op | Go 1.17 allocs/op | Go 1.26 allocs/op |")
    lines.append("|---|---:|---:|---:|---:|---:|---:|")
    for name in sorted(bench117):
        a = bench117[name]
        b = bench126[name]
        lines.append(
            f"| {name} | {a['ns/op']:.0f} | {b['ns/op']:.0f} | {a['B/op']:.0f} | {b['B/op']:.0f} | {a['allocs/op']:.0f} | {b['allocs/op']:.0f} |"
        )

    lines.append("")
    lines.append("### benchmark 关键结论")
    lines.append("")
    for name in sorted(bench117):
        a = bench117[name]
        b = bench126[name]
        b_reduction = (a["B/op"] - b["B/op"]) / a["B/op"] * 100
        alloc_delta = a["allocs/op"] - b["allocs/op"]
        lines.append(f"- {name}: B/op 下降约 {b_reduction:.1f}%，allocs/op 少 {alloc_delta:.0f} 次。")

    lines.append("")
    lines.append("## E4：size class / roundupsize 根因推导")
    lines.append("")
    lines.append("前提：`[]byte` 的元素大小为 1，因此 cap 基本等于申请字节数；实际 cap 会被 runtime 的 `roundupsize` 按 size class 向上取整。")
    lines.append("")
    lines.append("三步推导：")
    lines.append("")
    lines.append("1. Go 1.17 旧策略在 `oldcap < 1024` 时直接翻倍，所以 oldcap=1023 时理想 cap=2046，roundup 后得到 2048。")
    lines.append("2. 但 oldcap=1024 一跨过阈值，旧策略进入 1.25x 分支，理想 cap=1280，roundup 后仍是 1280。")
    lines.append("3. 因此 oldcap 从 1023 增加到 1024，append 后的新 cap 反而从 2048 掉到 1280。问题不是单独的 roundupsize，也不是单独的 1.25x，而是阈值硬切换 + size class 对齐共同制造了非单调。")
    lines.append("")
    lines.append("Go 1.26.2 中，oldcap=1023 和 oldcap=1024 append 后都得到 1536，说明这个断崖被平滑策略消掉了。")

    DATA.mkdir(parents=True, exist_ok=True)
    (DATA / "slice-growth-summary.md").write_text("\n".join(lines) + "\n")


if __name__ == "__main__":
    main()
