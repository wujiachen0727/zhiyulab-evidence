# Evidence 总索引

## 代码类

| 路径 | 说明 | 可运行 |
|------|------|:------:|
| code/split-comparison/monolith.go | 单体版"下单→扣库存" | ✅ |
| code/split-comparison/split.go | 拆分版"下单→扣库存" | ✅ |
| code/comm-comparison/main_test.go | 三种通信方式 benchmark | ✅ `go test -bench=. -benchmem` |

## 数据类

| 路径 | 说明 | 类型 |
|------|------|:----:|
| data/split-comparison-analysis.md | 单体 vs 拆分代码量对比分析 | 实测 |
| data/five-question-framework.md | 拆分决策5问检查法 | 推演 |
| data/communication-decision-tree.md | 通信选型决策树 | 推演 |

## 场景类

| 路径 | 说明 | 类型 |
|------|------|:----:|
| scenarios/three-teams.md | 三个团队的微服务决策故事 | 模拟 |

## 输出类

| 路径 | 说明 | 类型 |
|------|------|:----:|
| output/comm-comparison/benchmark-result.md | 通信方式 benchmark 结果 | 实测 |
