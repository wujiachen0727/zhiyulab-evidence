# Go 日志性能设计决策 Benchmark

配套文章：《Go 日志性能：5 个设计决策，比选库重要得多》

## 实验内容

| 实验 | 文件 | 验证目标 |
|------|------|---------|
| E1 | bench_file_test.go | 同步 vs Buffered 文件写入（4.8x 差异）|
| E2 | bench_test.go (Serialization) | JSON vs Text 序列化策略对比 |
| E3 | bench_test.go (Disabled Level) | disabled level 各库真实运行时开销 |
| E4 | bench_test.go (Field Binding) | 每次传 vs With 预绑定的性能差异 |
| E6 | bench_test.go (综合对比) | 同库默认 vs 优化配置 vs 跨库默认 |

## 运行方式

```bash
go test -bench=. -benchmem -count=3 -timeout=300s
```

## 环境要求

- Go >= 1.21（需要 log/slog 支持）
- 依赖：go.uber.org/zap, github.com/rs/zerolog

## 结果概要

- 同步直写文件 1465 ns/op vs Buffered 301 ns/op = 4.8x
- zap 默认 555 ns/op vs 优化 242 ns/op = 2.3x（同库）
- zap 默认 vs slog 默认 = 1.4x（跨库）
- zap typed disabled: 38 ns/op, 192B alloc（vs zerolog 3.9 ns, 0 alloc）
