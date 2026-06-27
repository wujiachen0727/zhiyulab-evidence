# E3 偶发毛刺证伪实验 — 最终结果汇总

**实验时间**：2026-04-19 08:00-08:06（多次迭代）
**Go 版本**：go1.26.2 (darwin/arm64)
**GOMAXPROCS**：14
**实测环境**：Apple Silicon M 系列芯片

---

## 证伪假设

**原假设（立意阶段）**：在 30 秒 pprof 采样窗口下，偶发毛刺会被**完全稀释**，pprof **看不到**。

**证伪结果**：**精确化成立**（不是简单的"成立"）。

---

## 实验设计

### 服务（main.go）
- `/fast`：快路径（~15μs CPU）——99% 流量
- `/slow`：在 15-20s 毛刺窗口内走慢路径（~200ms CPU），窗口外走快路径
- 独特函数名 `processSlowPath` 让它在 profile 里可识别

### 流量
- /fast：50 并发 × 1000 QPS → 实际 44598 req/s
- /slow：1 并发 × 10 QPS → 30s 共 23 次命中毛刺窗口

### 两组对比
- **Group A**（传统 pprof）：30 秒单窗口
- **Group B**（模拟持续 profiling）：每 5 秒一次 5 秒采样，共 6 个窗口

---

## 实验迭代史（诚实记录）

### Run 1（失败 1）：CPU profile 互斥

Group A 和 Group B 并发运行时，**Group B 6 个 curl 全部返回 "cpu profiling already in use"**。

根因：**Go 的 CPU profile 是全局单例的**。`runtime.StartCPUProfile` 同一时刻只能有一个实例。

**这本身是文章可用的一个点**：你不能"同时"跑传统 pprof 和持续 profiling——持续 profiling 工具（Pyroscope/Parca）用的是**完全不同的架构**（外部采样服务、时间切片聚合、eBPF 等）。

保留在 `output/run1-v1-too-light-cpu/` 作为证据。

### Run 2（失败 2）：慢路径太轻

拆分成两次独立运行后，Group A 和 B 都拿到了数据——但 `processSlowPath` **完全不出现**在任何 profile 里。

根因：**慢路径 CPU 工作量严重低估**。
- 预估：50-100ms
- 实测（/tmp/bench_paths）：500μs
- 在每秒 44598 QPS 的 HTTP 服务里，500μs × 250 次 = 125ms 总 CPU 时间——被 syscall 完全淹没

**教训**：硬件不同结果不同。Apple Silicon M 系列的计算速度远超预估，我按 x86 的经验估算严重偏小。

保留在 `output/run2-v1-too-light-cpu/` 作为证据。

### Run 3（成功）：慢路径加重 + 独立运行

- 慢路径工作量 × 400（8 亿次乘法），实测约 200ms（真实生产毛刺量级）
- Group A 独立运行（`run-experiment-part3.sh`）
- Group B 独立运行（`run-experiment-part2.sh`）

---

## 核心对比数据

### Group A：30 秒单窗口（传统 pprof）

```
Duration: 30.11s, Total samples = 96.48s (320.41%)
热点 TOP：
      flat    flat%   cum    cum%
    56.06s   58.11%   ...   syscall.rawsyscalln
    12.26s   12.71%   ...   runtime.usleep
     7.15s    7.41%   ...   runtime.kevent
     6.30s    6.53%   ...   runtime.pthread_cond_wait
     4.63s    4.80%   4.89s   5.07%  main.processSlowPath  ← 排第 5
     3.70s    3.83%   3.75s   3.89%  main.processFastPath  ← 排第 6
     3.54s    3.67%   ...   runtime.pthread_cond_signal
```

**关键发现**：
- `processSlowPath` **出现了**，但只有 **4.80%** flat（排第 5）
- `processFastPath` 占 3.83%——**两者混在一起，难以辨识**
- 你看到 4.80% 时，无法判断：
  - 是整 30 秒均匀发生的？还是集中在某 5 秒？
  - 4.80% 算异常还是正常背景？
  - 需要管吗？

### Group B：6 个 5 秒窗口（时间序列）

| 窗口 | 时间范围 | `processSlowPath` flat% | 备注 |
|:---:|:--------:|:----------------------:|:-----|
| 1 | 0-5s | — | 未出现（正常） |
| 2 | 5-10s | — | 未出现（正常） |
| 3 | 10-15s | 1.40% | 边缘泄漏（hey 请求时序不严格） |
| 4 | **15-20s** | **🔥 16.47%** | **毛刺主窗口！排第 2（仅次 syscall）** |
| 5 | 20-25s | — | 未出现（恢复） |
| 6 | 25-30s | — | 未出现（恢复） |

### 对比摘要

| 指标 | Group A 30s | Group B 窗口 4 | 差距 |
|------|:-----------:|:-------------:|:----:|
| `processSlowPath` flat% | 4.80% | **16.47%** | **3.43×** |
| 在 top 中的排名 | 第 5 | **第 2** | — |
| 能定位毛刺发生时段？ | ❌ | ✅ 15-20s | — |
| 能告诉你"其他时段正常"？ | ❌ | ✅ 窗口 1/2/5/6 全部 0% | — |
| 能判断毛刺是否异常？ | ❌（4.80% 看起来像背景） | ✅（16% vs 0% 差距明显） | — |

---

## 证伪结论（精确化）

**原假设**：30 秒采样稀释毛刺，pprof 看不到。
**实测精确化**：

> 30 秒单窗口下，毛刺会被**稀释**但**不会完全消失**——它以 4.80% 的占比排在第 5，混在 `processFastPath`（3.83%）和 runtime 调度（70%+）之间。
>
> 问题不是"看不到"，是"**看到了不知道怎么办**"。
>
> - 你不知道这 4.80% 是持续发生还是集中在某 5 秒
> - 你不知道 4.80% 是异常还是正常背景
> - 你的正常业务代码本就占 3.83%，毛刺和正常无法区分
>
> 时间序列让这一切一目了然——窗口 4 的 16.47% 对比其他窗口的 0%，**3.4 倍差距直接定位毛刺时间点**。持续 profiling 的价值不是"看得见"，是"**看得见时间**"。

这个修正比原假设更精确、更有洞察、更契合品牌"洞察、克制、可信"的调性。

---

## 意外附加发现（可用于文章）

1. **CPU profile 是全局单例的**（Run 1 发现）
   - 你不能同时跑"传统 pprof"和"持续 profiling"
   - 持续 profiling 工具的架构**必须**和 Go 原生 pprof 不同（外部采样、eBPF、时间切片）
   - 这解释了为什么 Pyroscope/Parca 不是"pprof 的扩展"，而是独立的基础设施

2. **毛刺在 top 里和正常业务无法区分**（Run 3 发现）
   - processSlowPath 4.80% vs processFastPath 3.83% —— 几乎没差距
   - 单看 flat% 根本无法判断哪个是异常
   - 这加强了"pprof 需要有对比基线"的论点——而持续 profiling 的 diff 功能正好解决这个

3. **Apple Silicon 的计算速度**（Run 2 发现）
   - 200 万次乘法在 M 系列芯片上只要 500μs（x86 经验值 5ms）
   - 写实验代码时不能按通用经验估算耗时——必须实测

---

## 对文章第 3 章立论的影响

**原立论**（立意阶段）："30 秒 pprof 采样稀释偶发毛刺，持续 profiling 能保留时间序列。"

**实测后立论**（更精准）：
> **30 秒 pprof 把毛刺变成了"底噪"——它在那儿，但你分不出来。**
>
> 我的实验里慢函数在 30s 窗口中占 4.80%——排在 top 第 5，就夹在正常业务代码之间。
>
> 但在 5 秒窗口的时间序列下：窗口 1、2、5、6 都是 0%，窗口 4 飙到 16.47%。时间轴告诉你"异常发生在第 15-20 秒"——这个信息不是 30 秒窗口"看得模糊一点"的问题，是**根本没有这个维度**。

这个立论用 **"底噪"** 这个比喻，比"稀释"更有画面感，也更精确。

---

## 实验产物清单

```
evidence/code/e3-spike/
├── main.go                         ← HTTP 服务（带 inSpikeWindow 逻辑）
├── run-experiment.sh               ← Run 1（失败：CPU profile 互斥）
├── run-experiment-part2.sh         ← Run 2（Group B 时间序列，成功）
├── run-experiment-part3.sh         ← Run 3（Group A 30s 单窗口，成功）
├── README.md                        ← 实验说明
└── output/
    ├── groupA-30s.pprof            ← Group A 最终成功数据（20 KB）
    ├── groupB-window-1.pprof       ← Group B 窗口 1-6 成功数据
    ├── groupB-window-{2..6}.pprof
    ├── groupB-timeseries/           ← Group B 数据的副本
    ├── run1-v1-too-light-cpu/       ← Run 1 失败记录（保留为证据）
    ├── run2-v1-too-light-cpu/       ← Run 2 失败记录（保留为证据）
    ├── server-part3.log             ← 服务日志
    ├── hey-fast-part3.log / hey-slow-part3.log  ← 压测日志
    └── final-stats-part3.txt        ← 最终请求统计
```

### 可读文本输出
```
evidence/output/e3-spike/
├── e3-summary.md                    ← 本文件
└── （后续生成 top.txt 系列）
```

---

## 下一步

- ✅ E3 证伪完成
- ✅ 立意精确化（"底噪"隐喻 + 3.43× 数字对比）
- 🎯 下一个实验：**E5**（Pyroscope 本地部署 + 开销实测）——或者 E1/E4（pprof 基线）
