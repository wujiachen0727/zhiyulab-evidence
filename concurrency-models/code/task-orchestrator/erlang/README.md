# Erlang/Actor 风格任务编排器实验

## 目的

用 Erlang process + message + link 表达同一个任务编排器，观察 Actor 心智模型下三类责任的位置：

1. 状态所有权：每个 worker process 拥有自己的局部状态，parent 只聚合消息。
2. 调度责任：parent 通过 mailbox `receive ... after` 定义等待边界。
3. 失败边界：worker 退出以 `EXIT` 信号进入 parent，parent 决定是否 kill 兄弟进程。

## 运行

```bash
escript orchestrator.escript
```

## 输出说明

输出包含 `erlang-success`、`erlang-timeout`、`erlang-worker-error` 三个场景。本实验只观察责任表达位置，不代表完整 OTP supervision 设计。
