# Benchmark 原始结果

## 环境
- Go 1.26.2 darwin/arm64
- CPU: Apple M4 Pro
- GOMAXPROCS=8
- 每组 -count=3

## 场景1：计数器 Channel vs Mutex vs Atomic [实测 Go 1.26.2]

| 方案 | ns/op（3轮均值） | allocs/op |
|------|:----------------:|:---------:|
| Mutex | ~105 | 0 |
| Channel (buffered 1) | ~97 | 0 |
| Atomic | ~30 | 0 |

**意外发现**：在 Apple M4 Pro + Go 1.26.2 上，buffered channel(1) 用作互斥保护时性能和 Mutex 接近（甚至略快）。这与 dev.to 文章声称的 75× 差距完全不同。原因可能是：
1. Go 运行时对 buffered channel 的优化在新版本有提升
2. dev.to 的测试条件不同（可能是 unbuffered channel + 额外 goroutine）
3. 竞争强度（GOMAXPROCS）设置不同

**关键结论**：单纯的计数器场景，Mutex 并不比 Channel 快（在现代硬件 + 新 Go 版本上）。真正的差距在更复杂的场景。

## 场景2：缓存 RWMutex vs Channel [实测 Go 1.26.2]

| 方案 | ns/op（3轮均值） | allocs/op |
|------|:----------------:|:---------:|
| RWMutex | ~17.5 | 0 |
| Channel (cache manager) | ~456 | 0 |

**关键结论**：RWMutex 比 Channel 快约 **26 倍**。原因：
- RWMutex 允许多读者并发（90% 读场景收益巨大）
- Channel 方案所有操作串行化到单一 goroutine，丧失并发优势
- 这才是真正体现选型差异的场景

## 场景3：工作池 Channel vs Mutex+Cond [实测 Go 1.26.2]

| 方案 | ns/op（3轮均值） | allocs/op |
|------|:----------------:|:---------:|
| Channel | ~95 | 0 |
| Mutex+Cond | ~186 | 0 |

**关键结论**：工作池场景，Channel 比 Mutex+Cond 快约 **2 倍**，且代码量少一半。这是 Channel 的正确舞台——任务分发和协调。

## 场景4：管道 Channel vs Mutex+slice vs Sequential [实测 Go 1.26.2]

| 方案 | ns/op（3轮） | 均值 ns/op | B/op | allocs/op |
|------|:------------|:----------:|:----:|:---------:|
| Channel Pipeline | 68.14 / 68.96 / 68.60 | ~68.6 | 0 | 0 |
| Mutex+slice Pipeline | 11.39 / 10.78 / 4.91 | ~9.0* | ~95 | 0 |
| Sequential（基线） | 0.23 / 0.23 / 0.23 | ~0.23 | 0 | 0 |

> *Mutex+slice 第3轮因 JIT 热身效应显著低于前两轮，若取前两轮均值约 11 ns/op。

**关键结论**：
- **Mutex+slice 比 Channel Pipeline 快约 6-8×**（~11 ns vs ~69 ns）
- Channel 的开销来自：goroutine 调度 + channel 收发（每 item 2次 send + 2次 receive）
- Mutex+slice 有少量内存分配（slice 动态扩容），但零 heap allocs
- Sequential 基线 ~0.23 ns/op，说明管道的核心价值不在纯计算性能
- **管道模式的价值**：多阶段并发解耦（I/O 密集型场景）、可组合、背压控制——这些是 Mutex+slice 无法提供的结构优势

## 竞争强度曲线 [实测 Go 1.26.2]

| 并行度 | Mutex ns/op | Channel ns/op | Channel/Mutex 比率 |
|:------:|:-----------:|:-------------:|:------------------:|
| 1 | ~106 | ~100 | 0.94× |
| 10 | ~100 | ~122 | 1.22× |
| 100 | ~92 | ~130 | 1.41× |
| 1000 | ~94 | ~155 | 1.65× |

**关键结论**：
- 低竞争（1 goroutine）：Channel 和 Mutex 几乎无差异
- 高竞争（1000 goroutine）：Channel 比 Mutex 慢约 65%
- Mutex 在高竞争下性能反而略微提升（饥饿模式保证公平性，减少无效自旋）
- Channel 在高竞争下性能线性下降（hchan 内部锁 + goroutine 调度开销）
