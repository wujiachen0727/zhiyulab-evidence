import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ARTICLE_EVIDENCE = ROOT.parents[1]
OUTPUT_DIR = ARTICLE_EVIDENCE / "output" / "deliverable-gap"
DATA_DIR = ARTICLE_EVIDENCE / "data"


def load_module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def non_empty_code_lines(path):
    return sum(1 for line in path.read_text(encoding="utf-8").splitlines() if line.strip())


def run_checks(module, mode):
    checks = []

    def record(name, passed, detail):
        checks.append({"name": name, "passed": bool(passed), "detail": detail})

    # 1. 正常路径
    try:
        if mode == "initial":
            result = module.calculate_checkout({"items": [{"price": 10, "qty": 2}], "coupon": "VIP"})
        else:
            result = module.calculate_checkout({"items": [{"price": 10, "qty": 2}], "coupon": "VIP"}, request_id="req-1")
        record("正常路径可计算", result.get("total") in (16.0, "16.00"), str(result))
    except Exception as exc:
        record("正常路径可计算", False, repr(exc))

    # 2. 空购物车必须拒绝
    try:
        if mode == "initial":
            module.calculate_checkout({"items": []})
        else:
            module.calculate_checkout({"items": []}, request_id="req-empty")
        record("空购物车被拒绝", False, "未报错")
    except Exception as exc:
        record("空购物车被拒绝", True, type(exc).__name__)

    # 3. 负价格必须拒绝
    try:
        if mode == "initial":
            module.calculate_checkout({"items": [{"price": -5, "qty": 1}]})
        else:
            module.calculate_checkout({"items": [{"price": -5, "qty": 1}]}, request_id="req-neg")
        record("负价格被拒绝", False, "未报错")
    except Exception as exc:
        record("负价格被拒绝", True, type(exc).__name__)

    # 4. 数量为 0 必须拒绝
    try:
        if mode == "initial":
            module.calculate_checkout({"items": [{"price": 10, "qty": 0}]})
        else:
            module.calculate_checkout({"items": [{"price": 10, "qty": 0}]}, request_id="req-zero")
        record("零数量被拒绝", False, "未报错")
    except Exception as exc:
        record("零数量被拒绝", True, type(exc).__name__)

    # 5. request_id / 幂等追踪
    try:
        if mode == "initial":
            result = module.calculate_checkout({"items": [{"price": 10}]})
            record("幂等追踪字段存在", "request_id" in result, str(result))
        else:
            module.calculate_checkout({"items": [{"price": 10}]}, request_id="")
            record("幂等追踪字段存在", False, "空 request_id 未报错")
    except Exception as exc:
        record("幂等追踪字段存在", mode != "initial", type(exc).__name__)

    # 6. 审计事件
    try:
        if mode == "initial":
            result = module.calculate_checkout({"items": [{"price": 10}]})
        else:
            result = module.calculate_checkout({"items": [{"price": 10}]}, request_id="req-audit")
        record("审计事件存在", bool(result.get("audit_events")), str(result))
    except Exception as exc:
        record("审计事件存在", False, repr(exc))

    # 7. 货币单位
    try:
        if mode == "initial":
            result = module.calculate_checkout({"items": [{"price": 10}]})
        else:
            result = module.calculate_checkout({"items": [{"price": 10}]}, request_id="req-currency")
        record("货币单位明确", result.get("currency") == "CNY", str(result))
    except Exception as exc:
        record("货币单位明确", False, repr(exc))

    # 8. 折扣配置可替换
    try:
        if mode == "initial":
            result = module.calculate_checkout({"items": [{"price": 100}], "coupon": "TEST"})
            record("折扣规则可配置", result.get("total") == 90.0, str(result))
        else:
            result = module.calculate_checkout(
                {"items": [{"price": 100}], "coupon": "TEST"},
                request_id="req-rules",
                discount_rules={"TEST": module.Decimal("0.10")},
            )
            record("折扣规则可配置", result.get("total") == "90.00", str(result))
    except Exception as exc:
        record("折扣规则可配置", False, repr(exc))

    return checks


def summarize(checks):
    passed = sum(1 for item in checks if item["passed"])
    return {"passed": passed, "total": len(checks), "failed": len(checks) - passed}


def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    DATA_DIR.mkdir(parents=True, exist_ok=True)

    initial_path = ROOT / "ai_initial.py"
    deliverable_path = ROOT / "deliverable_version.py"
    initial = load_module(initial_path, "ai_initial")
    deliverable = load_module(deliverable_path, "deliverable_version")

    initial_checks = run_checks(initial, "initial")
    deliverable_checks = run_checks(deliverable, "deliverable")

    result = {
        "environment": "Python stdlib, no external dependencies",
        "note": "ai_initial.py 是本次论证阶段由 AI 按最小可运行需求生成的初版；deliverable_version.py 是按可交付护栏补齐后的版本。",
        "code_lines": {
            "ai_initial": non_empty_code_lines(initial_path),
            "deliverable_version": non_empty_code_lines(deliverable_path),
        },
        "checks": {
            "ai_initial": initial_checks,
            "deliverable_version": deliverable_checks,
        },
        "summary": {
            "ai_initial": summarize(initial_checks),
            "deliverable_version": summarize(deliverable_checks),
        },
    }

    (OUTPUT_DIR / "result.json").write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")

    initial_summary = result["summary"]["ai_initial"]
    deliverable_summary = result["summary"]["deliverable_version"]
    initial_failed = [item for item in initial_checks if not item["passed"]]

    markdown = [
        "# 从能跑到可交付的缺口统计",
        "",
        "[实测 Python 3.9.6 stdlib] 本实验比较同一个结账计算功能的 AI 初版与可交付版本。",
        "",
        "## 结果摘要",
        "",
        f"- AI 初版非空代码行：{result['code_lines']['ai_initial']} 行。",
        f"- 可交付版本非空代码行：{result['code_lines']['deliverable_version']} 行。",
        f"- AI 初版通过护栏：{initial_summary['passed']} / {initial_summary['total']}。",
        f"- 可交付版本通过护栏：{deliverable_summary['passed']} / {deliverable_summary['total']}。",
        f"- 从能跑到可交付，至少补齐 {deliverable_summary['passed'] - initial_summary['passed']} 个护栏项。",
        "",
        "## AI 初版缺失项",
        "",
    ]

    for item in initial_failed:
        markdown.append(f"- {item['name']}：{item['detail']}")

    markdown.extend([
        "",
        "## PR 审查清单对比",
        "",
        "| 护栏项 | AI 初版 | 可交付版本 |",
        "|---|:---:|:---:|",
    ])

    for left, right in zip(initial_checks, deliverable_checks):
        markdown.append(
            f"| {left['name']} | {'通过' if left['passed'] else '缺失'} | {'通过' if right['passed'] else '缺失'} |"
        )

    markdown.extend([
        "",
        "## 论证用途",
        "",
        "这个实验不证明所有 AI 代码都不可靠；它只证明一个更窄的工程事实：第一版“能跑”的代码，通常还没有携带足够的交付证据。严肃工程的价值，不是拖慢探索，而是把这些证据补齐到可以审查、回滚和追责。",
        "",
    ])

    (DATA_DIR / "deliverable-gap-summary.md").write_text("\n".join(markdown), encoding="utf-8")
    print(json.dumps(result["summary"], ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
