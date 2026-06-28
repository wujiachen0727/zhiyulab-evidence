# E2 三层对比实验

## 实验目的

对比三种方案的 P99、服务端请求量、mutex 等待时间：
- A：无优化基线（单锁 + 无 hedging）
- B：仅 hedging（单锁 + hedging）——症状治疗
- C：修复锁竞争 + hedging（分片锁 + hedging）——病因治疗 + 症状兜底

## 运行方法

```bash
go run main.go
```

## 环境

- Go 1.26.4 darwin/arm64
