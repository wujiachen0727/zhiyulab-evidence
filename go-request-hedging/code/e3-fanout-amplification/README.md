# E3 Fan-out 放大实测

## 实验目的

验证《The Tail at Scale》的数学推导：单节点 1% 超时概率 × 100 并行节点 = 63% 整体超时概率。

## 运行方法

```bash
go run main.go
```

## 环境

- Go 1.26.4 darwin/arm64
