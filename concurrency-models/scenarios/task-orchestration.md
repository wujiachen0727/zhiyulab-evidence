# 场景模拟：聚合 3 个下游服务返回用户画像

## 场景

一个用户画像接口要同时拉取三个下游：

1. `profile`：基础资料，通常最快。
2. `billing`：付费状态，偶尔返回业务错误。
3. `risk`：风控标签，偶尔超时。

接口目标不是“谁跑得最快”，而是回答三个工程问题：

- 状态归谁：聚合结果由谁持有，worker 能不能直接改？
- 等待归谁：谁负责等待慢下游，谁负责超时？
- 失败归谁：一个子任务失败后，谁决定取消兄弟任务？

## 三种模型下的观察

| 模型 | 状态所有权 | 调度责任 | 失败边界 |
|------|------------|----------|----------|
| Go / CSP 风格 | 聚合 goroutine 持有结果，worker 通过 channel 回传 | `context` / `select` 显式表达等待与取消 | 第一个错误触发 cancel，兄弟任务通过 context 收到取消 |
| Erlang / Actor 风格 | worker process 拥有局部状态，parent 聚合消息 | parent mailbox 的 `receive ... after` 定义等待边界 | worker exit 变成 parent 可观察信号，parent 决定 kill 兄弟进程 |
| Java virtual thread 风格 | 调用方普通对象持有聚合状态 | 阻塞写法保留，等待成本由 JVM virtual thread 承接 | 应用层 Future 编排决定 timeout/error 后取消哪些任务 |

## 正文可用结论

同一个“拉三个下游再聚合”的任务，三种模型真正不同的不是语法，而是责任露在哪里：Go 把协作关系露在 channel/context 上，Erlang 把边界露在 process/message/exit 上，Java virtual thread 把同步代码保留下来，但失败边界仍要应用自己定义。
