# OTel 采样与内存（文献 + 工程共识）

## 背景

OpenTelemetry Go SDK 在 high-traffic 场景下，100% head sampling 可能导致 RSS 持续增长（参见 open-telemetry/opentelemetry-go Discussion #7748）。

## 推荐基线（待论证阶段引用，非本地复现）

| 采样策略 | 预期 trace 量 | 内存风险 |
|---------|-------------|---------|
| ParentBased + TraceIDRatioBased(0.1) | ~10% | 低 |
| ParentBased + TraceIDRatioBased(1.0) | 100% | 高（高 QPS 下易 OOM） |
| BatchSpanProcessor 默认队列 2048 | — | 队列满则丢 span，不阻塞业务 |

## 正文引用策略

- 引用 OTel 官方：采样是 trace 性能首要杠杆
- 不编造具体 RSS 数字——若无本地 benchmark，正文写「高 QPS 下全采样有 OOM 报告，建议从 10% ParentBased 起步」
- 若后续补 benchmark，写入 `evidence/output/otel-memory/`

标注：**当前为工程共识整理；RSS 对比数字不在正文出现，除非补跑实验。**
