# Evidence 总索引

> 文章：《写 Go linter 不难，难的是让团队用起来》
> 实测环境：Go 1.26.2 darwin/arm64

## 实验代码

| 目录 | 实验 | 状态 | 关键结论 |
|------|------|:----:|---------|
| code/func-length-linter/ | E7 最简 linter | ✅ 实测通过 | go/analysis 框架核心逻辑 ~20 行 |
| code/error-check-compare/ | E1 go/ast vs go/types 误报对比 | ✅ 实测通过 | go/ast 12条检测 vs go/types 4条（噪音减少 67%）|
| code/architecture-linter/ | E2 架构分层 linter | ✅ 实测通过 | go/ast 检出 1/3 vs go/types 3/3（go/ast 漏掉接口调用和 alias 调用）|

## 实测数据

| 文件 | 内容 | 类型 |
|------|------|:----:|
| data/golangci-lint-integration.md | E3 golangci-lint 集成踩坑 | 推演+文档溯源 |
| data/team-adoption-simulation.md | E5 团队引入三阶段 | 场景模拟 |

## 实验输出

| 目录 | 内容 |
|------|------|
| output/func-length-linter/ | E7 运行结果 |
| output/error-check-compare/ | E1 运行结果（对比表格）|
| output/architecture-linter/ | E2 运行结果（对比表格）|

## 论据自造度统计

- 独立论据 7 项：E1-E7 全部自造
- 实验验证 4 项（E1/E2/E3/E7）：3 项实测通过，1 项推演+文档溯源
- 数据实测 1 项（E4）：基于 E1 实验数据提取
- 场景模拟 1 项（E5）：已标注为场景模拟
- 逻辑推演 1 项（E6）：直接融入正文
- **自造比例：7/7 = 100%**
- 外部引用：2 处（go/analysis 官方文档、golangci-lint 文档），非论据性
