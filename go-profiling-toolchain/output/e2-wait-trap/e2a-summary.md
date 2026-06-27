# E2a 锁争用实验最终输出

**实验时间**：2026-04-19 07:55:41
**Go 版本**：go1.26.2 (darwin/arm64)
**GOMAXPROCS**：14
**实验配置**：50 workers × 200000 tasks × critSize=5000（每个临界区 ~6-10μs）
**实验耗时**：1.38 秒
**完成任务**：200,049 个
**平均每任务**：6.92μs

---

## 三种 pprof 视角 + Trace 对比

### 视角 1：CPU Profile

```
Duration: 1.50s, Total samples = 1980ms (131.84%)  ← 14 核并发采样
热点：
- runtime.pthread_cond_signal    65.66%   1300ms
- runtime.pthread_cond_wait      24.24%   480ms
- runtime.usleep                 10.10%   200ms

其他全是 runtime.findRunnable / runtime.schedule / runtime.mcall
main.incrementHeavy 完全不出现
```

**能告诉你**：CPU 时间花在调度器让 goroutine 入睡/唤醒上
**不能告诉你**：为什么要睡？睡了多久？谁在等谁？

---

### 视角 2：Mutex Profile

```
Showing nodes accounting for 64.05s, 100% of 64.05s total
      flat  flat%   cum%
    64.05s   100%   100%  sync.(*Mutex).Unlock
         0     0%   100%  main.(*shared).incrementHeavy
         0     0%   100%  main.runWorkload.func1
```

**反直觉发现**：100% 归因给 `sync.Mutex.Unlock`——不是 Lock。
- **原因**：Go mutex profile 记录"你释放锁时导致了多少累积等待"——把因果反向归到 Unlocker
- 不理解这个设计的人会以为"Unlock 很慢"——这是真实存在的误读陷阱

**能告诉你**：锁争用的累积时长（64.05 秒）
**不能告诉你**：什么时候争用？哪些 goroutine 在争？争用分布是否均匀？

---

### 视角 3：Block Profile

```
Showing nodes accounting for 71.82s, 100% of 71.82s total
      flat  flat%   cum%
    66.75s 92.94%  92.94%  sync.(*Mutex).Lock (inline)
     2.30s  3.20%  96.15%  runtime.chanrecv1
     1.38s  1.93%  98.07%  runtime.selectgo
```

**能告诉你**：谁在 Lock 上等了多久（66.75 秒）；谁在 channel 上等（2.30 秒）
**不能告诉你**：等待的时间序列、goroutine 之间的因果、唤醒顺序

---

### 视角 4：Execution Trace

```
总事件数: 803,875
总 goroutine 数: 75
Goroutine 状态转移：
  Runnable → Running    176,355 次
  Running → Waiting     119,271 次
  Waiting → Runnable    119,267 次
  Running → Runnable     57,033 次（被抢占/主动让出）

每个 goroutine 详细分布（TOP 20）：
goid     waitMs     runMs      runnableMs   nWait    nRun
18       1385       0          0            2        1       （等 WaitGroup）
23       1350       19         13           1841     3007    
75       1344       21         17           2177     3335    
80       1339       22         21           2517     3671    
...

汇总：
总阻塞 wait time:    83,673 ms  （≈ 50 goroutine × 1.4s × 高阻塞比）
总运行 run time:      1,441 ms  （墙钟 1.4s，几乎只有 1 个核在跑）
总可运行 runnable:      994 ms  （"想跑但抢不到 CPU"的时间）
```

**Trace 独家能给**：
- 每个 goroutine 的 wait/run/runnable 时间分布
- 每次阻塞/调度的时间戳（精度 ns 级）
- goroutine 被阻塞了**多少次**（worker 平均阻塞 2000 次以上）
- "runnable but not running"（可运行但抢不到 CPU）的时间——**pprof 完全没有这个维度**

---

## 核心数据对比表（给文章用）

| 维度 | CPU profile | Mutex profile | Block profile | Trace |
|------|:----------:|:-------------:|:-------------:|:-----:|
| 能看见 CPU 热点 | ✅ 但全是 runtime 调度 | ❌ | ❌ | ✅ |
| 能看见锁等待总量 | ❌ | ✅ 64.05s | ✅ 66.75s | ✅ 83.67s |
| 锁等待归因点 | N/A | Unlock（反直觉） | Lock | 每次具体事件 |
| 能看见 channel 阻塞 | ❌ | ❌ | ✅ 2.3s | ✅ |
| 能看见"抢不到 CPU"的等待 | ❌ | ❌ | ❌ | ✅ 0.99s |
| 能看见单个 goroutine 分布 | ❌（混合采样） | ❌ | ❌ | ✅ |
| 能看见时间序列 | ❌ | ❌ | ❌ | ✅ |
| 能看见阻塞次数 | ❌ | ❌ | ❌ | ✅（平均 2000+ 次） |
| 能看见因果关系 | ❌ | ❌ | ❌ | ✅ |
| 文件大小 | 1.2 KB | 1.9 KB | 1.5 KB | 3.6 MB |
| 生产开启成本 | 低（持续） | 很低 | 很低 | 高（通常只短期） |

---

## 证伪结论

**原假设**：pprof 无法识别等待导致的慢。
**证伪结果**：**部分成立**（需要精确化表述）。

**精确化结论**：
1. **CPU profile 看不见等待**（完全成立） ✅
   - 证据：50 worker 锁争用 1.38 秒，CPU profile 只显示 runtime 调度 idle，**业务代码完全不可见**
2. **Block/Mutex profile 能看见等待的聚合总量**（原假设不成立） ⚠️
   - 证据：Block profile 正确显示 66.75 秒锁等待
3. **但 pprof 全家桶都看不见等待的时间序列、次数分布、因果关系**（新假设成立） ✅
   - 证据：Trace 显示平均每个 worker 阻塞 2000 次以上——pprof 只给"总共 66.75 秒"一个数
   - 证据：Trace 给出"runnable 0.99 秒"——pprof 没有这个维度

---

## 对文章第 2 章立论的影响

**原立论（粗糙）**："pprof 看不见等待，所以你需要 trace。"

**实测后的精确立论（更有洞察）**：
> pprof 能告诉你**累计等了多久**，trace 才能告诉你**什么时候等、等了多少次、跟谁有关**。
> 
> 你拿到一个"锁争用 66.75 秒"的数字，不知道：
> - 是 1 次等了 66 秒，还是 66 万次各等 100μs？（trace 告诉你是 **11 万次**）
> - 是所有 goroutine 均匀等，还是 1% 的 goroutine 吃掉 90% 的等待？（trace 告诉你分布）
> - 这 66 秒里，有多少是"想跑但抢不到 CPU"？（trace 告诉你约 0.99 秒）
> 
> pprof 给统计——trace 给结构。这是层次差异，不是替代关系。

这个观点**比原立论更精准、更反直觉、更有"洞察"感**——完美契合"止语Lab"品牌。

---

## 不再需要 E2b

原计划 E2b 是"纯 channel 阻塞"实验——但 E2a 已经同时覆盖了 channel 阻塞（block profile 2.3s）。再做 E2b 会重复。

**E2 最终产出**：
- `run1-original/`：第一次失败的实验（channel 主导，锁没被争起来）——保留作为**反例教训**（告诉读者"实验设计很容易翻车"）
- `e2a-mutex-contention/`：成功的证伪实验（锁争用 + channel 混合）

这两个加起来**本身就是一个完整的叙事**——"我第一次实验设计错了，数据告诉我 channel 成了节流阀，重做之后才看到真相"。这种叙事在"锐利观点"风格下**非常契合**，也避免了只贴成功结果的"AI 腔"。

---

## 下一步

- ✅ E2 证伪完成，立意精确化调整已完成
- 🎯 下一个实验：**E3 偶发毛刺证伪**（证明 30s 窗口稀释偶发毛刺，持续 profiling 能留下时间序列）
- 可选：**E1 + E4**（CPU 热点基线 + 采样频率对比）
- 可选：**E5**（Pyroscope 本地部署 + 开销实测）
