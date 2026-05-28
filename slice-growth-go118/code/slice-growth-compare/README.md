# slice-growth-compare

用于对比 Go 1.17 与当前 Go 版本的 `[]byte` slice 扩容行为。

## 环境

- Go 1.17.13：通过 `go install golang.org/dl/go1.17.13@latest && go1.17.13 download` 安装
- 当前 Go：本机 `go version`
- 测试机器：darwin/arm64

## 运行方式

```bash
# 安装旧版 Go 包装器后，确保 go1.17.13 在 PATH 中
# go install golang.org/dl/go1.17.13@latest && go1.17.13 download

# 容量增长序列
go1.17.13 run . -mode append -n 5000
go run . -mode append -n 5000

# 从 cap=1024 开始增长
go1.17.13 run . -mode grow-from-cap -start-cap 1024 -n 5000
go run . -mode grow-from-cap -start-cap 1024 -n 5000

# benchmark
go test -bench=. -benchmem -count=5
go1.17.13 test -bench=. -benchmem -count=5
```

## 注意

- 本实验是实测 runtime 行为，不是算法复现。
- benchmark 使用 package-level sink 变量，避免 DCE 优化。
- 数据解读优先看 `allocs/op` 和 `B/op`，再看 `ns/op`。
