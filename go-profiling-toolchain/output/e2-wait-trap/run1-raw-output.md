# E2 第一次运行的原始输出

**实验时间**：2026-04-19 07:53:06
**Go 版本**：go1.26.2 (darwin/arm64)
**GOMAXPROCS**：14
**实验耗时**：1.00 秒（4019 任务完成）

---

## 原始控制台输出

```
=== E2 等待陷阱证伪实验 ===
Go 版本: go1.26.2, GOMAXPROCS: 14, NumCPU: 14
✅ 完成任务数: 4019, 耗时: 1.001475958s
✅ 全部 profile 已写入 output/ 目录
```

---

## CPU Profile（`go tool pprof -top`）

```
File: main
Type: cpu
Time: 2026-04-19 07:53:06 CST
Duration: 1.10s, Total samples = 100ms ( 9.07%)
Showing nodes accounting for 100ms, 100% of 100ms total
      flat  flat%   sum%        cum   cum%
      60ms 60.00% 60.00%       60ms 60.00%  runtime.kevent
      40ms 40.00%   100%       40ms 40.00%  runtime.pthread_cond_wait
         0     0%   100%      100ms   100%  runtime.findRunnable
         0     0%   100%       40ms 40.00%  runtime.mPark (inline)
         0     0%   100%      100ms   100%  runtime.mcall
         0     0%   100%       60ms 60.00%  runtime.netpoll
         0     0%   100%       40ms 40.00%  runtime.notesleep
         0     0%   100%      100ms   100%  runtime.park_m
         0     0%   100%      100ms   100%  runtime.schedule
         0     0%   100%       40ms 40.00%  runtime.semasleep
         0     0%   100%       40ms 40.00%  runtime.stopm
```

### 关键数据
- **Duration 1.10s，Total samples 100ms（9.07%）** —— 1 秒墙钟时间里，20 个 worker（14 核机器）总共只用了 100ms CPU
- **热点全部在 runtime 调度器的 idle 等待**（`kevent`, `pthread_cond_wait`, `netpoll`, `semasleep`）
- **业务代码完全不出现**（没有 `main.cpuWork`, `main.incrementWithLock`, `main.blockOnChannel`）

---

## Block Profile（`go tool pprof -top`）

```
File: main
Type: delay
Showing nodes accounting for 23.96s, 100% of 23.96s total
Dropped 12 nodes (cum <= 0.12s)
      flat  flat%   sum%        cum   cum%
    21.96s 91.64% 91.64%     21.96s 91.64%  runtime.chanrecv1
        1s  4.18% 95.82%         1s  4.18%  runtime.selectgo
        1s  4.18%   100%         1s  4.18%  sync.(*WaitGroup).Wait
         0     0%   100%     19.96s 83.29%  main.blockOnChannel (inline)
         0     0%   100%         1s  4.18%  main.main
         0     0%   100%         1s  4.18%  main.runWorkload
         0     0%   100%     19.96s 83.29%  main.runWorkload.func1
```

### 关键数据
- **23.96 秒累计阻塞时间**（20 个 worker × 1 秒墙钟 ≈ 20 秒理论上限——接近满阻塞）
- **91.64% 归因给 `chanrecv1`**（channel 接收阻塞）
- **19.96 秒归因给 `main.blockOnChannel`**（占总阻塞 83.29%）
- **没有锁争用出现在 block profile 里**（说明 sync.Mutex 没有触发显著 contention）

---

## Mutex Profile（`go tool pprof -top`）

```
File: main
Type: delay
Showing nodes accounting for 0, 0% of 0 total
(空结果)
```

### 关键数据
- **Mutex profile 为空**
- **深层原因**（raw 查看后确认）：本次实验的 `sync.Mutex` 没有真正产生争用——因为 20 个 worker 大部分时间卡在 `blockOnChannel`（200μs 级）上，到达 Mutex 的速度不均匀，实际同时竞争锁的瞬间极少

---

## Trace（文件大小：199927 字节）

未在本次提取可读输出。后续分析需启动 `go tool trace` Web UI。

---

## 诊断：为什么锁争用没发生

**锁临界区耗时估算**：
- 1024 次 `s.data[i] = ...` + `s.counter++` ≈ 300-500ns（L1 cache 写）

**锁外耗时估算**：
- `cpuWork`（50 次乘法）≈ 50ns
- `blockOnChannel`（等下游 200μs+）= 200,000ns 起

**比例**：锁临界区占 worker 单次迭代的 0.15-0.25%，channel 阻塞占 99%+。
→ **channel 成了天然的"节流阀"**，20 个 worker 很少同时冲向锁。

---

## 本次实验的重要启示

1. **假设 1**（CPU profile 看不见等待）：**完全成立** ✅
   - 1 秒墙钟只有 100ms CPU 采样，且全在 runtime 调度器的 idle 等待
   - 业务代码在 CPU profile 里**完全不可见**

2. **假设扩展**：block profile 能看见 channel 阻塞 ✅
   - block profile 完整记录了 19.96s 的 channel 阻塞时间
   - **这和原本"pprof 看不见等待"的粗糙说法有矛盾**——pprof 的 block profile 能看见一部分等待

3. **对立意的精确化修正**：
   - 原说法："pprof 看不见等待，trace 才看得见"——**太粗糙**
   - 精确说法："pprof 的 CPU/block/mutex profile 都在回答'累计在哪'；trace 在回答'什么时候/为什么/跟谁一起'"
   - 这是一个**更反直觉、更有洞察**的观点

---

## 实验需要重构

**E2 拆分为 E2a + E2b**：

- **E2a**：纯锁争用场景（移除 channel 阻塞）→ 让 mutex profile 有数据
- **E2b**：纯 channel 阻塞场景 → 证明 block profile 能看见
- **核心对比**：即使 pprof 能看见等待（通过 block/mutex profile），它给的也只是**聚合统计**——trace 给的是**时间序列 + 调度关系 + 因果链**

这个版本的立意会比原版更精准、更独特。
