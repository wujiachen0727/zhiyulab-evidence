# Evidence 总索引：slice-growth-go118

本文核心论点：Go 1.18 的 slice 扩容改动不是单纯优化，而是在修复旧策略的隐藏问题：非单调增长、阈值跳变和增长曲线断裂。

---

## 环境预检

| 项 | 结果 | 说明 |
|---|:---:|---|
| 当前 Go | ✅ | `go version go1.26.2 darwin/arm64` |
| Go 1.17.13 | ✅ | 已通过 `go install golang.org/dl/go1.17.13@latest && go1.17.13 download` 安装 |
| Python 3 | ✅ | 用于汇总分析输出 |

---

## 论据清单

| ID | 类型 | 自造/引用 | 状态 | 产出路径 | 正文使用建议 |
|---|------|:--------:|:---:|---------|-------------|
| E1 | 实验验证 | 自造 | ✅ 完成 | `evidence/code/slice-growth-compare/`、`evidence/output/slice-growth-compare/go117-scan-900-1400.csv`、`evidence/data/slice-growth-summary.md` | 开场和第 1 章使用：oldcap 1023→newcap 2048；oldcap 1024→newcap 1280 |
| E2 | 数据实测 | 自造 | ✅ 完成 | `evidence/output/slice-growth-compare/go117-append-5000.csv`、`go126-append-5000.csv`、`evidence/data/slice-growth-summary.md` | 第 1/3 章使用：Go 1.17 与 Go 1.26 容量序列对比 |
| E3 | 数据实测 | 自造 | ✅ 完成 | `evidence/output/slice-growth-compare/go117-bench.txt`、`go126-bench.txt`、`evidence/data/slice-growth-summary.md` | 第 4 章使用：B/op 下降约 12.2%-14.5%，allocs/op 少 1 次 |
| E4 | 逻辑推演 | 自造 | ✅ 完成 | `evidence/data/slice-growth-summary.md` | 第 2 章使用：阈值硬切换 + size class 对齐共同制造非单调 |
| E5 | 场景模拟 | 自造 | ✅ 完成 | `evidence/scenarios/interview-code-review.md` | 收尾使用：面试/code review 中的过时八股代价 |
| E6 | 外部引用 | 引用 | ⏳ 留给 grounding | 待补：Go 官方 commit / runtime 源码 | 第 3 章只作佐证，不作为核心论据 |

---

## 关键结论

### E1：旧策略非单调增长确实存在

`[实测 Go 1.17.13 darwin/arm64]`

扫描 oldcap=900..1400，append 1 个 byte 后：

- oldcap=1023 → newcap=2048
- oldcap=1024 → newcap=1280

也就是说，旧 cap 只增加 1，append 后的新 cap 反而下降 768。Go 1.26.2 同范围下降点为 0。

### E2：1.18+ 后增长曲线更平滑

`[实测 Go 1.17.13 / Go 1.26.2 darwin/arm64]`

1024 附近窗口：

| oldcap | Go 1.17 newcap | Go 1.26 newcap |
|---:|---:|---:|
| 1023 | 2048 | 1536 |
| 1024 | 1280 | 1536 |
| 1025 | 1408 | 1536 |

旧策略在 1024 阈值处断崖式下降；新策略保持平滑。

### E3：benchmark 显示分配更少、B/op 更低

`[实测 Go 1.17.13 / Go 1.26.2 darwin/arm64]`

| benchmark | B/op 变化 | allocs/op 变化 |
|---|---:|---:|
| From1024To4096 | 12544 → 11008（下降约 12.2%） | 5 → 4 |
| NoPrealloc_4K | 14584 → 12536（下降约 14.0%） | 13 → 12 |
| Prealloc256_4K | 14080 → 12032（下降约 14.5%） | 7 → 6 |

注意：本文不引用 CSDN 的 17-40% 数据，使用自测结果替代。

### E4：根因不是单一公式，而是两阶段交互

前提：`[]byte` 元素大小为 1，cap 基本等于申请字节数；最终 cap 会被 `roundupsize` 按 size class 向上取整。

三步推导：

1. Go 1.17 在 `oldcap < 1024` 时直接翻倍，所以 oldcap=1023 时理想 cap=2046，roundup 后得到 2048。
2. oldcap=1024 一跨过阈值，旧策略进入 1.25x 分支，理想 cap=1280，roundup 后仍是 1280。
3. 因此 oldcap 从 1023 增加到 1024，append 后的新 cap 从 2048 掉到 1280。

结论：问题不是单独的 `roundupsize`，也不是单独的 1.25x，而是阈值硬切换 + size class 对齐共同制造了非单调。

---

## 自造比例

- 独立论据总数：6 项
- 自造论据：5 项（E1-E5）
- 外部引用：1 项（E6，待 grounding）
- 自造占比：83.3%

去掉 E6 后，文章核心观点仍成立。
