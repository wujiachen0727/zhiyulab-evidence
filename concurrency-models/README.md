# Evidence 索引：并发模型三流派

## 执行摘要

本阶段完成 8 项自造论据中的 8 项，外部引用 3 项仅作为事实校准。核心自造证据是同一个“任务编排器”在 Go、Java virtual thread、Erlang/Actor 风格下的可运行实现，以及 timeout / worker-error 场景输出。

## 环境

详见 `output/environment-preflight.md`：

- Go：go1.26.2 darwin/arm64
- Java：OpenJDK 26.0.1
- Erlang：Erlang/OTP 29

## 自造论据清单

| ID | 类型 | 状态 | 产出路径 | 正文用途 |
|----|------|:---:|----------|----------|
| E1 | 实验验证 | ✅ 完成 | `code/task-orchestrator/{go,java,erlang}/` + `output/task-orchestrator/*-output.txt` | 同一任务编排器在三种模型下的代码形态差异 |
| E2 | 实验验证 | ✅ 完成 | `output/task-orchestrator/summary.md` | timeout / worker-error 下失败边界如何扩散 |
| E3 | 场景模拟 | ✅ 完成 | `scenarios/task-orchestration.md` | 聚合 3 个下游服务的统一场景 |
| E4 | 逻辑推演 | ✅ 完成 | `scenarios/model-boundaries.md` | 三问框架与“默认心智模型 ≠ 能力边界” |
| E5 | 数据实测 | ✅ 完成 | `data/responsibility-surface.md` | 责任显性化辅助统计，不作性能指标 |
| E6 | 场景模拟 | ✅ 完成 | `scenarios/incident-debugging.md` | 线上排障视角解释三种模型的排查入口 |
| E7 | 逻辑推演 | ✅ 完成 | `scenarios/model-boundaries.md` | 回应 Go/Java/Erlang 都能写其他模型的反例 |
| E8 | 场景模拟 | ✅ 完成 | `scenarios/decision-cheatsheet.md` | 结尾选型速查表 |

## 外部引用清单

| ID | 状态 | 用途 | 已有记录 |
|----|:---:|------|----------|
| R1 | ✅ 已求证 | Go goroutine/channel 与“通过通信共享内存”事实校准 | `drafts/grounding-log.md` |
| R2 | ✅ 已求证 | Erlang process/message/link/monitor 事实校准 | `drafts/grounding-log.md` |
| R3 | ✅ 已求证 | Java virtual thread 官方定位事实校准 | `drafts/grounding-log.md` |

## 自造度统计

- 自造独立论据：8 项
- 外部引用：3 项
- 自造占比：8 / (8 + 3) = 72.7%
- 结论：达到目标 ≥ 70%

## 正文引用边界

1. 任务编排器实验只支撑“责任表达位置不同”，不支撑性能优劣。
2. `responsibility-surface.md` 是代理指标，正文必须明确说明不能推出语言优劣。
3. Java 实验使用 virtual thread 基础 API，不能写成 StructuredTaskScope 实测。
4. Erlang 实验是最小 process/message/link 表达，不等同完整 OTP supervision。
