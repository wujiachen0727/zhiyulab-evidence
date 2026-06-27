# map-sync-bench

## 目的

实测普通 map、`map + sync.RWMutex`、`sync.Map` 在相同读写比例下的单 goroutine 操作开销，用来说明：普通 map 默认不加锁不是疏忽，而是避免让所有场景支付同步成本。

## 运行环境

- Go：go1.26.2 darwin/arm64
- 运行命令：`go test -bench=. -benchmem -count=5`

## 说明

- 普通 map 仅作为单 goroutine 基线，不代表并发安全方案。
- `ReadMostly`：约 99% 读、1% 写。
- `WriteHeavy`：约 50% 读、50% 写。
- benchmark 使用 package-level sink 变量接收结果，避免 DCE 优化。
