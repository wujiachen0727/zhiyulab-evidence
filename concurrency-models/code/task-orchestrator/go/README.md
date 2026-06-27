# Go 任务编排器实验

## 目的

用同一个“并发拉取 profile / billing / risk 三个下游并聚合结果”的场景，观察 Go 风格代码把三类责任放在哪里：

1. 状态所有权：聚合状态由主 goroutine 持有，worker 只通过 channel 发送不可变结果。
2. 调度责任：等待、超时和取消由 `context.WithTimeout` 与 `select` 显式表达。
3. 失败边界：第一个 worker 错误触发 `cancel()`，通过共享 context 传播给兄弟任务。

## 运行

```bash
go run main.go
```

## 输出说明

输出包含 3 个场景：

- `go-success`：三个下游都成功。
- `go-timeout`：`risk` 超过整体超时，被 context 取消。
- `go-worker-error`：`billing` 返回错误，编排器取消其他任务。

本实验不测性能，只观察“责任在代码表面如何显性化”。
