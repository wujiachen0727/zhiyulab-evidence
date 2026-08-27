# E6 goroutineleak 检出 vs 漏检

Go 1.27.0，darwin/arm64。

## 可检出：局部 channel 阻塞

`local-channel`：goroutine 阻塞在函数内 `make(chan struct{})` 上。

```
Showing nodes accounting for 1, 100% of 1 total
```

profile 大小 448 bytes。

## 漏检：全局 channel 阻塞

`global-channel`：goroutine 阻塞在包级 `var globalCh` 上。

```
Showing nodes accounting for 0, 0% of 0 total
```

profile 大小 231 bytes（空 profile）。

与 release notes 一致：阻塞在**经全局变量仍可达**的并发原语上，GC 可达性分析可能**判不成泄漏**。

代码：`evidence/code/goroutineleak/main.go`
