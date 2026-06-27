# Evidence 总索引

## 论据清单

| ID | 类型 | 类别 | 描述 | 产出路径 | 状态 |
|----|------|:----:|------|---------|:----:|
| E1 | 实验验证 | 独立论据 | Go testing vs JUnit 5 API 面积对比 | evidence/data/api-compare.md | ✅ |
| E2 | 数据实测 | 独立论据 | Go 标准库 Test/Bench 比例统计 | evidence/data/stdlib-test-ratio.md | ✅ |
| E3 | 实验验证 | 独立论据 | assert 摩擦成本实测 | evidence/data/assert-friction.md + evidence/code/assert-friction/ | ✅ |
| E4 | 场景模拟 | 独立论据 | 缓存模块 TDD vs Go 惯用对照 | evidence/scenarios/cache-module.md | ✅ |
| E5 | 逻辑推演 | 独立论据 | 设计哲学推导 testing 必然性 | 融入正文 | ✅ |
| E6 | 逻辑推演 | 独立论据 | testify 使用率 | evidence/data/testify-usage.md | ⚠️ 降级 |
| E7 | 逻辑推演 | 独立论据 | 并发测试确定性对比 | evidence/data/concurrent-test.md | ⚠️ 降级 |
| R1 | 外部引用 | — | Russ Cox "Go Testing By Example" | 行内引用 | 待正文 |
| R2 | 外部引用 | — | Dijkstra 测试哲学 | 行内引用 | 待正文 |

## 统计

- 自造论据：7 项 / 总 9 项 = 78%
- 外部引用：2 处（上限 3 处）
- 降级项：E6（数据实测→推演，原因：精确数据不可获取）、E7（实验验证→推演，原因：多版本 Go 环境不可用）
