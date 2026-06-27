# Evidence 总索引

**文章**：为什么你的 Go TCP server P99 延迟这么高
**生成时间**：2026-04-26

## 实验列表

| ID | 类型 | 描述 | 状态 | 路径 |
|----|------|------|:----:|------|
| E1 | 数据实测 | TCP_NODELAY 单向批量发送对比（64B/512B/4KB x 10000） | ✅ 完成 | 复用 go-network-programming/evidence |
| E2a | 数据实测 | TCP_NODELAY 多网络环境对比（本地/模拟1ms/模拟10ms） | ✅ 完成 | `code/nodelay-compare/` → `output/` |
| E2b | 数据实测 | TCP_NODELAY 请求-响应模式对比（32B-4KB） | ✅ 完成 | `code/nodelay-reqresp/` → `output/` |
| E3 | 实验验证 | SO_RCVBUF/SO_SNDBUF 调优对比（8KB-1MB） | ✅ 完成 | `code/buffer-tuning/` → `output/` |
| E4 | 源码分析 | Go 标准库默认 socket 选项 | ✅ 完成 | 融入正文（Go 源码 net 包） |
| E5 | 经验落地 | 诊断叙事骨架（排查故事线） | ✅ 融入正文 | — |
| E6 | 逻辑推演 | "什么时候该调"判断框架 | ✅ 融入正文 | — |
| E7 | 逻辑推演 | Nagle + Delayed ACK 互锁原理 | ✅ 融入正文 | 标注 [推演]，macOS 不支持 TCP_QUICKACK 无法实测 |

## 关键数据摘要

### E1: TCP_NODELAY 单向批量发送（原实测数据复用）
- 小包(64B) x 10000: NoDelay=true 慢 206%
- 原因：禁用 Nagle 后系统调用次数暴增

### E2b: TCP_NODELAY 请求-响应
- 小请求(32B-64B)：差异 <10%（本地回环 RTT≈0）

### E3: 缓冲区调优
- 系统默认 SO_RCVBUF≈400KB, SO_SNDBUF≈147KB
- 8KB 缓冲区吞吐量降 59%（238 vs 578 MB/s）
- 256KB/1MB 与默认差异不大

## 环境信息

- Go 版本：go1.26.2
- OS/Arch：darwin/arm64
- CPU：Apple M4 Pro
