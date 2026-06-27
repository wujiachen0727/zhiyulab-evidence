# 任务编排器实验摘要

## 实测范围

同一个聚合任务被实现为三种风格：

1. Go：goroutine + channel + context。
2. Java：virtual thread + CompletionService/Future。
3. Erlang：process + message + link/EXIT。

每种实现都运行 3 个场景：

- success：三个下游都成功。
- timeout：`risk` 下游超过整体超时。
- worker-error：`billing` 下游返回错误。

## 输出文件

| 语言 | 输出 |
|------|------|
| Go | `evidence/output/task-orchestrator/go-output.txt` |
| Java | `evidence/output/task-orchestrator/java-output.txt` |
| Erlang | `evidence/output/task-orchestrator/erlang-output.txt` |

## 可用于正文的观察

| 模型 | 状态所有权 | 调度责任 | 失败边界 |
|------|------------|----------|----------|
| Go / CSP 风格 | 聚合 goroutine 持有结果，worker 只发 `Result` | `context.WithTimeout` 和 `select` 显式表达等待/取消 | 第一个错误触发 `cancel()`，兄弟任务通过 context 感知取消 |
| Java virtual thread | 调用方普通对象持有聚合状态 | 阻塞式 Future 编排保留，等待成本由 JVM virtual thread 承接 | timeout/error 后仍需应用层取消未完成 Future |
| Erlang / Actor 风格 | worker process 拥有局部状态，parent 聚合消息 | parent mailbox 的 `receive ... after` 定义等待边界 | worker exit 变成 parent 可观察信号，parent kill 兄弟进程 |

## 使用边界

- 本实验不是性能 benchmark。
- 数据只支撑“责任表达位置不同”，不能推出三种模型谁更快或谁绝对更先进。
- Java 实验使用 virtual thread 基础 API，没有使用 preview/孵化中的结构化并发 API；正文应避免说成 StructuredTaskScope 实测。
- Erlang 实验是最小 Actor 风格表达，不等同于完整 OTP supervision 工程。
