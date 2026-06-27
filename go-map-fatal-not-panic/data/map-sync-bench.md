# E2 map 同步成本 benchmark

## 环境

- Go：go1.26.2 darwin/arm64
- CPU：Apple M4 Pro
- 命令：`go test -bench=. -benchmem -count=5`
- 原始输出：`evidence/output/map-sync-bench/bench.txt`
- 实验代码：`evidence/code/map-sync-bench/map_sync_bench_test.go`

## 测量范围说明

普通 map 只作为单 goroutine 基线，不代表并发安全方案。本文比较的是「默认不加锁」与「显式同步结构」的操作开销量级，不把它写成生产环境绝对性能结论。

`ReadMostly` 约 99% 读、1% 写；`WriteHeavy` 约 50% 读、50% 写。所有 benchmark 使用 package-level sink 变量接收结果，避免 DCE 优化。

## 平均结果（5 次）

| 场景 | 实现 | 平均 ns/op | B/op | allocs/op | 相对普通 map |
|---|---|---:|---:|---:|---:|
| ReadMostly | plain map | 3.86 | 0 | 0 | 1.00x |
| ReadMostly | map + RWMutex | 4.85 | 0 | 0 | 1.25x |
| ReadMostly | sync.Map | 15.46 | 0 | 0 | 4.00x |
| WriteHeavy | plain map | 5.76 | 0 | 0 | 1.00x |
| WriteHeavy | map + RWMutex | 6.57 | 0 | 0 | 1.14x |
| WriteHeavy | sync.Map | 33.36 | 31 | 1 | 5.79x |

## 可供正文引用的结论

在这组本机实测里，给普通 map 外面套 `RWMutex` 后，单 goroutine 基准下读多写少场景从约 3.86 ns/op 到 4.85 ns/op，写多场景从约 5.76 ns/op 到 6.57 ns/op；`sync.Map` 在写多场景约 33.36 ns/op，并出现 1 alloc/op。

这个数据不能推出「sync.Map 一定慢」；它只支撑一个更克制的判断：并发安全不是免费的。Go 不让普通 map 默认带同步，是为了不让所有普通使用场景都支付这笔成本。
