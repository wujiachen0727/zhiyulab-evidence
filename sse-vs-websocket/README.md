# 论据总索引

> **文章**：WebSocket 是个好东西，但你不需要它——从 AI 流式到实时推送，SSE 的逆袭
> **生成阶段**：Argue
> **生成日期**：2026-06-11

---

## 自造论据

| # | 类型 | 文件 | 优先级 | 状态 | 正文引用 |
|---|------|------|:------:|:----:|:--------:|
| E1 | 经验落地 | `scenarios/E1-experience-ws-overuse.md` | 1 | ✅ 完成 | 第1章 |
| E2 | 逻辑推演 | `scenarios/E2-protocol-overhead.md` | 6 | ✅ 完成 | 第2章 |
| E3 | 数据实测 | `code/benchmark/benchmark_test.go` → `output/benchmark-results.txt` | 2 | ✅ 完成 | 第2章、第3章 |
| E4 | 逻辑推演（降级） | `scenarios/E4-http2-multiplexing.md` | 7 | ✅ 完成（降级：环境限制） | 第4章 |
| E5 | 场景模拟 | `scenarios/E5-scenario-ws-misuse.md` | 4 | ✅ 完成 | 第1章 |
| E6 | 经验落地 | `scenarios/E6-production-pitfalls.md` | 5 | ✅ 完成 | 第6章 |
| E7 | 逻辑推演 | `scenarios/E7-decision-framework.md` | 3 | ✅ 完成 | 第5章 |

## 外部引用

| # | 内容 | 来源 | 引用位置 |
|---|------|------|:--------:|
| R1 | 浏览器兼容性数据 | MDN / caniuse.com | 第4章 |
| R2 | WebSocket 协议规范 | RFC 6455 / websocket.org | 第2章 |
| R3 | AI 厂商流式 API | OpenAI/Anthropic 官方文档 | 第3章 |

## 统计数据

- 自造论据：7 项
- 外部引用：3 项
- 自造占比：7/10 = **70%** ✅（≥ 70% 目标）
- 降级处理：1 项（E4：实验验证→逻辑推演，降级原因：环境限制）

## Benchmark 关键数据

运行 `go test -bench=. -benchmem` 在 Apple M4 Pro / Go 1.26.4 环境：

| 指标 | SSE | WebSocket | 差异 |
|------|:---:|:---------:|:----:|
| 耗时（ns/op） | 1,125,961,167 | 2,324,352,208 | WS 慢 106% |
| 内存分配（B/op） | 75,732,024 | 93,550,288 | WS 多 23% |
| 分配次数（allocs/op） | 552,563 | 730,270 | WS 多 32% |

**结论**：在 1000 并发单向推送场景下，SSE 的耗时约为 WebSocket 的一半，内存分配少约 23%。
