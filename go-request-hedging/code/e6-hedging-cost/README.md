# E6 hedging 成本实测

## 实验目的

测量 hedging 的隐藏成本：
- 连接池压力（额外请求量）
- 内存分配和 GC 压力
- 可观测性污染（日志、metrics、trace）

## 运行方法

```bash
go run main.go
```

## 环境

- Go 1.26.4 darwin/arm64
