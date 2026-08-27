# 跨服务超时链预算（推演 + 模型）

## 模型

- 调用链深度：3 层（Gateway → Order → Inventory）
- 每层 HTTP/gRPC client timeout：3s
- 串行最坏等待：3 × 3s = **9s**（非 3s——常见误解）
- 峰值并发 200 RPS，若下游 Inventory 变慢，Gateway 层同时挂起的 goroutine 上限 ≈ min(并发请求, 200×9s 窗口内堆积)

## 推演结论

| 指标 | 数值 | 说明 |
|------|------|------|
| 单层 timeout | 3s | 团队常见默认值 |
| 三层串行最坏 | 9s | 每层独立计时 |
| 200 并发全慢 | 200 goroutine × ~9s 占用 | 仅 Gateway 进程 |
| 建议 client timeout | ≤ 500ms–1s（内网） | 留 budget 给重试/排队 |
| 建议 context 总 deadline | 2s（用户-facing） | 小于各层 timeout 之和 |

标注：**算术推演，非压测实测。** 论证阶段用逻辑推演支撑「超时链乘法效应」子论点。
