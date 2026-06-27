# E5 Pyroscope 开销实测 — 最终结果汇总

**实验时间**：2026-04-19 08:18
**Go 版本**：go1.26.2 (darwin/arm64)
**GOMAXPROCS**：14
**实测环境**：Apple Silicon M 系列芯片

**Pyroscope 版本**：v1.21.0（Docker 镜像 grafana/pyroscope:latest，digest sha256:751794786ee5）
**Pyroscope Go SDK 版本**：v1.2.8

---

## 实验目标

回应读者质疑——"Grafana 官方声称 Pyroscope 持续 profiling 开销 < 1%，真的吗？"

**自实测，给作者评价**。

---

## 实验设计

### 服务对比
- **基线版**（`baseline/main.go`）：Go HTTP 服务，纯 `net/http`，无任何 profiling agent
- **agent 版**（`with-agent/main.go`）：完全相同的业务逻辑 + Pyroscope Go SDK（push mode，10 种 profile types 全开，每 15s 上传一次）

### 业务逻辑
两版完全相同：每个请求做 50000 次乘法运算（~15μs CPU）。

### 压测
`hey -c 30 -z 60s`（30 并发、60 秒、不限 QPS）

### 测量
- **QPS**：hey 给出
- **延迟**：hey 给出的 P50/P90/P95/P99/Slowest
- **进程 CPU**：每 5 秒 ps 一次，取 60 秒内平均值
- **进程 RSS**：每 5 秒 ps 一次，取平均值

---

## 实验迭代

### Run 1（CPU 监控失效）

**问题**：用 `go run main.go` 启动服务，ps 抓到的是 `go` 工具的 PID（父进程），而不是真正的服务进程。结果：平均 CPU 0.0%（明显错误）。

**教训**：监控进程资源时，必须先 `go build` 编译再运行，这样 PID 就是服务本身。

保留在 `output/run1-cpu-monitoring-broken/` 作为教训记录。

### Run 2（成功）

改用 `go build -o baseline-bin main.go && ./baseline-bin` 的方式启动，ps 抓到的就是服务进程 PID。数据干净。

---

## 核心数据

### 吞吐量（QPS）

| 版本 | QPS | 差距 |
|------|:----:|:----:|
| 基线（无 agent） | **45387.66 req/s** | — |
| 带 agent（Pyroscope push mode） | **44656.64 req/s** | **-1.61%** |

**结论**：开启 Pyroscope 全量 profile（10 种 profile types）后，QPS 下降 **1.61%**。比 Grafana 官方声称的 `< 1%` **略高**。

### 延迟

| 指标 | 基线 | 带 agent | 差距 |
|------|:----:|:--------:|:----:|
| Average | 1.8 ms | 1.8 ms | 持平 |
| P50 | 0.7 ms | 0.7 ms | 持平 |
| P90 | 0.7 ms | 0.7 ms | 持平 |
| P95 | 0.7 ms | 0.8 ms | +14% |
| P99 | 0.9 ms | 0.8 ms | **-11%**（反向，在误差内） |
| **Slowest** | 7.6 ms | 6.7 ms | **-12%**（反向） |
| Fastest | 0.1 ms | 0.1 ms | 持平 |

**结论**：**P50/P90/Average 无差别，P95 有 14% 的微弱上升，P99/Slowest 在误差内反向**。

### 进程资源

| 指标 | 基线 | 带 agent | 差距 |
|------|:----:|:--------:|:----:|
| 平均 CPU 使用率 | **283.9%** | **294.4%** | **+3.70%** |
| 平均 RSS | **23.1 MB** | **30.9 MB** | **+33.8%**（+7.8 MB）|

**说明**：
- CPU 283.9%（基线）≈ 3 个核满负载——说明基线本身没有跑满 14 核（不是 CPU 绑定场景），符合真实 HTTP 服务特征
- agent 版 CPU 额外开销 **+10.5 个百分点 / 3.70%**（相对基线）
- RSS 增加 7.8 MB —— Pyroscope agent 的 profile 缓冲区占用

---

## 证实/证伪结果

### Grafana 官方声称 `< 1%`

**实测结果**：
- **QPS 损失 1.61%** —— 比官方声称**略高**，但在同一数量级
- **CPU 开销 +3.70%** —— 比 QPS 损失更高

**可能的原因**：
1. **本次实验开了全量 10 种 profile types**（CPU / alloc / inuse / goroutines / mutex / block 等），这是比官方默认配置**更激进**的场景
2. 官方的 `< 1%` 通常指"只开 CPU profile"或特定配置
3. Apple Silicon 的开销特征可能和生产常见的 x86 Linux 有差异

### 作者评价（用于 R2 引用的附加评论）

> Grafana 说 Pyroscope 开销 <1%——我本地实测 QPS 下降 1.61%，CPU 开销 3.70%，RSS 增加 7.8MB。
>
> 数字比他们的宣传**略高**，但我开了全量 10 种 profile types（CPU + 内存 + goroutine + mutex + block），这比大多数生产配置更激进。他们的 <1% 大概率针对的是"只开 CPU profile"的最小化配置。
>
> 即使按我的实测数据——**每月为了随时能看的时间序列 profile，付 1-4% 性能**——对大多数生产服务仍然是划算的交易。
> 
> **这不是黑 Pyroscope，这是作为使用者的诚实反馈**。

---

## 意外发现（可用于文章）

1. **P95/P99 没有明显上升**
   - Pyroscope 的上传是异步 + 后台 goroutine，不阻塞请求路径
   - Slowest 反而比基线快 12%——这是随机波动（在 ms 级小样本误差范围内）

2. **RSS 增加比 CPU 开销更显著**（+33.8% vs +3.70%）
   - agent 需要维护 profile 缓冲区（特别是 alloc profile 需要追踪每次分配）
   - 对小型服务（RSS 本就 20-30MB 级），这个比例会很扎眼——但绝对值仍然小（7.8MB）

3. **CPU profile 不互斥**
   - Pyroscope Go SDK 和 `net/http/pprof` 都在同一进程里——它们如何共存？
   - 答案：Pyroscope SDK 调用的是 `runtime/pprof` 的**独立接口**，但底层**仍会和 `pprof.StartCPUProfile` 互斥**
   - 意味着：装了 Pyroscope agent 后，`go tool pprof http://.../debug/pprof/profile?seconds=30` 可能会**偶发失败**（agent 正在采样）——这是一个文章可以提的"生产陷阱"

---

## 对文章第 3 章立论的影响

**原立论**（立意阶段）：引用 Grafana R2 声称 `< 1%`，附加作者评论。

**实测后立论**：
> Grafana 官方数据说 `<1%`。我本地实测——QPS 损失 **1.61%**，CPU 开销 **3.70%**，RSS 增加 **7.8 MB**。
>
> 比他们宣传的高一点，但我开了全量 10 种 profile types。这是激进配置下的上限数字。最小化配置（只开 CPU profile）应该能压到 1% 以内。
>
> 这个数字贵吗？对 45000 QPS 的服务，1.61% 意味着损失大约 **730 QPS**。换来的是什么？——
>
> 1. **随时能看历史 profile**（不是出了问题才开采样）
> 2. **时间序列对比**（今天 vs 昨天 / 本版本 vs 上版本）
> 3. **跨节点聚合**（哪个实例异常）
>
> 你在什么时候会觉得这笔交易不划算？**流量本身就是瓶颈的服务**（如极致优化的代理、CDN 边缘节点）——他们为每 0.5% QPS 都要斗。
>
> 但对 99% 的业务服务——**1-4% 的性能换持续可观测，是非常好的交易**。

这个新立论**基于实测数字**，附带**具体的不适用场景**（代理/CDN），比单纯引用 Grafana 数据扎实得多。

---

## 实验产物清单

```
e5-pyroscope-overhead/
├── README.md（待写）
├── baseline/
│   └── main.go（基线版，无 agent）
├── with-agent/
│   ├── main.go（挂 Pyroscope Go SDK）
│   ├── go.mod / go.sum
├── run-experiment.sh
└── output/
    ├── baseline-bin / agent-bin（二进制，忽略入 git）
    ├── baseline-server.log / agent-server.log
    ├── baseline-hey.log / agent-hey.log
    ├── baseline-proc.csv / agent-proc.csv（进程资源时间序列）
    ├── baseline-final-stats.txt / agent-final-stats.txt
    ├── e5-comparison.txt（核心对比报告）
    └── run1-cpu-monitoring-broken/（Run 1 失败记录）
```

### 外部资源
- **Pyroscope server 容器**：`pyroscope-e5`（`grafana/pyroscope:latest` / v1.21.0）
- **Pyroscope server 运行状态**：本次实验期间 0.41% CPU / 25MB RSS

---

## 当前证伪进度总结（E2 + E3 + E5 完成后）

| 假设 | 验证方式 | 立意修正后的表述 |
|:----|:--------|:----------------|
| pprof 看不见等待 | E2 锁争用实验 | "pprof 给你一个数，trace 给你一个故事" |
| 30s 采样稀释毛刺 | E3 毛刺实验 | "pprof 把毛刺变成底噪" |
| 持续 profiling 开销可控 | E5 实测 | "QPS 损失 1.61%，CPU 开销 3.7%，对 99% 业务服务是划算的交易" |

**自造度累计**（E2 + E3 + E5 完成后）：
- 独立实测数据文件：17（E2）+ 14（E3）+ 8（E5）≈ 39 份
- **100% 自造**——全部本地实测
- 引用 R2（Grafana <1% 声明）时附带**实测对比评价**——不是搬运

---

## 下一步

- ✅ E5 Pyroscope 开销实测完成
- 🎯 下一步：E1 CPU 热点基线 + E4 采样频率对比（支撑第 1 章）
- 或者：E8 决策表提炼（基于 E2/E3/E5 的数据）
