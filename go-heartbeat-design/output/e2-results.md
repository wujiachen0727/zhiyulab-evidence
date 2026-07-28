# E2: 心跳风暴 + Jitter 缓解实验结果

## 实验环境
- macOS Darwin arm64
- Go 1.26.2 (via goenv)

## 实验设计
- 10000 个连接，心跳间隔 30s
- 无 Jitter：所有连接用相同 30s 间隔
- 有 Jitter：每个连接用 30s ± 5s 随机偏移
- 测量一次 tick 的耗时和内存占用

## 结果

| 场景 | 一次 tick 耗时 | HeapInuse |
|------|:------------:|:---------:|
| 无 Jitter | 116ms | 13.09 MB |
| 有 Jitter (30s±5s) | 206ms | 6.86 MB |

## 分析
- **内存**：Jitter 使内存占用降低 47%（13.09→6.86 MB），因为 goroutine 启动时间分散，不会同时分配
- **耗时**：Jitter 场景耗时略长（206ms vs 116ms），因为 tick 时间分散在更大窗口内
- **结论**：Jitter 有效缓解心跳风暴，内存峰值降低近一半
