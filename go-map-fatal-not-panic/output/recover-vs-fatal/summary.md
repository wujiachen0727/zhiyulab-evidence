# E1 recover vs fatal 实验结果

## 环境

- Go：go1.26.2 darwin/arm64
- 机器：Apple Silicon（darwin/arm64）
- 实验代码：`evidence/code/recover-vs-fatal/main.go`

## 运行结果

| 模式 | 命令 | exit code | stdout/stderr 关键输出 | 结论 |
|---|---|---:|---|---|
| 普通 panic | `recover-vs-fatal panic` | 0 | `recovered ordinary panic: ordinary panic` | 同 goroutine 内 `defer recover` 可恢复普通 panic |
| concurrent map writes | `recover-vs-fatal fatal` | 2 | `fatal error: concurrent map writes` | runtime fatal 直接终止进程，`defer recover` 没有恢复机会 |

## 可供正文引用的结论

我跑了一个最小对比：普通 `panic` 可以被同 goroutine 的 `defer recover` 接住，进程 exit code 为 0；但并发写 map 触发的是 runtime fatal，stderr 第一行就是 `fatal error: concurrent map writes`，进程 exit code 为 2。

这说明：`recover` 恢复的是用户态 panic 控制流，不是 runtime 对不可信数据结构的终止判定。
