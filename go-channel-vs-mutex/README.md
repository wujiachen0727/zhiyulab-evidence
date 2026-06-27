# Evidence 总索引

## 环境

- Go 1.26.2 darwin/arm64
- CPU: Apple M4 Pro
- GOMAXPROCS=8

## 论据清单

| ID | 类型 | 描述 | 状态 | 产出路径 |
|----|------|------|:----:|---------|
| E1 | 实验验证 | 4 种场景 benchmark（计数器/缓存/工作池/管道） | ✅ 完成 | `code/benchmark/*_test.go` → `output/benchmark-results.md` |
| E2 | 场景模拟 | Channel 滥用导致 goroutine 泄漏 | ✅ 完成 | `code/goroutine-leak/main.go` |
| E3 | 数据实测 | 不同竞争强度性能曲线（1/10/100/1000 并行度） | ✅ 完成 | `output/benchmark-results.md` § 竞争强度曲线 |
| E4 | 逻辑推演 | Mutex 饥饿模式：正常→饥饿切换机制 | ✅ 完成 | 融入正文 |
| 引用 | 外部引用 | Go Proverbs 原文 | ✅ 使用 | 行内引用 + 作者评价 |

## 关键发现

### 意外发现：计数器场景 Mutex ≈ Channel

dev.to 文章声称 Mutex 比 Channel 快 75 倍。我的实测（Go 1.26.2, M4 Pro）显示：

- 计数器场景：Mutex ~105ns vs Channel ~97ns，几乎无差异
- **真正差距出现在缓存场景**：RWMutex ~17.5ns vs Channel ~456ns（26 倍）

原因：buffered channel(1) 用作互斥锁时，开销和 Mutex 接近。但当场景需要读写分离（RWMutex）时，Channel 的"全串行化"劣势暴露无遗。

### Channel 的正确舞台

- 工作池：Channel ~95ns vs Mutex+Cond ~186ns（Channel 快 2 倍）
- 管道：Channel 的价值在结构而非性能

### 竞争强度影响

Mutex 在高竞争下性能反而微升（饥饿模式减少无效自旋），Channel 线性下降。

## 统计

- 自造论据：4 项（E1 实验验证 + E2 场景模拟 + E3 数据实测 + E4 逻辑推演）
- 外部引用：1 处（Go Proverbs）
- 自造度：4/5 = **80%** ✅
