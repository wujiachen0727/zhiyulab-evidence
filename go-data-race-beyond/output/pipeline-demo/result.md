# E5 实验输出 — pipeline 所有权模型对比

**运行命令**：`go run -race main.go`
**Go 版本**：1.26.4 darwin/arm64
**运行时间**：2026-06-19

## 结果

| 版本 | Data Race | 结果完整性 | 说明 |
|------|:---------:|:----------:|------|
| 无所有权 | **5 个** | 丢失 3 个元素（2/5） | 每个阶段共享 slice，两次 append 都有 data race |
| 有所有权（channel pipeline） | **0 个** | 完整（5/5） | producer → processor → consumer 链式传递所有权 |

## 关键结论

- Pipeline 模式下，无所有权设计会导致两个阶段的 data race
- 有所有权设计：每个 goroutine 只拥有当前阶段的数据，通过 channel 传递后释放
- 这是 Go 中 channel pipeline 模式的正确打开方式
