# recover-vs-fatal

## 目的

对比普通 `panic` 与 `fatal error: concurrent map writes` 的恢复行为。

## 运行环境

- Go：go1.26.2 darwin/arm64

## 运行方式

```bash
go build -o ../../output/recover-vs-fatal/recover-vs-fatal .
../../output/recover-vs-fatal/recover-vs-fatal panic
../../output/recover-vs-fatal/recover-vs-fatal fatal
```

## 判定标准

- `panic` 模式：同 goroutine 内的 `defer recover` 应捕获普通 panic，进程 exit code 为 0。
- `fatal` 模式：并发写 map 触发 runtime throw，defer/recover 不会执行恢复，进程以非 0 exit code 退出，stderr 包含 `fatal error: concurrent map writes` 或相近 runtime fatal 信息。
