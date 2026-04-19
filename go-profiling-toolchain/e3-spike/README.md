# E3：偶发毛刺证伪实验

## 证伪假设

**原假设（立意阶段）**：30 秒 pprof 采样窗口稀释偶发毛刺——pprof 看不到，持续 profiling 能。

**证伪结果**：**精确化成立**。更精确的表述是：

> 30 秒单窗口下，毛刺**变成了底噪**——排在 top 第 5（4.80%），混在正常业务代码（3.83%）之间，你分不出来是毛刺还是背景。时间序列下，窗口 4 的 16.47% 对比其他窗口的 0%，**3.4 倍差距**直接定位到 15-20s 时段。

## 实验迭代（诚实记录）

### Run 1 - CPU profile 互斥（失败但有价值）

**发现**：Go 的 CPU profile 是全局单例。Group A（30s 采样）和 Group B（6 × 5s 采样）并发时，Group B 全部返回 `cpu profiling already in use`。

**文章可用的点**：这证明了为什么持续 profiling 工具必须用**完全不同的架构**（外部采样、eBPF、时间切片）——而不是"pprof 的扩展"。

保留证据：`output/run1-v1-too-light-cpu/`

### Run 2 - 慢路径太轻（失败的估算）

**发现**：Apple Silicon M 系列芯片上 200 万次乘法只要 500μs（x86 经验值约 5ms），慢路径被 syscall 完全淹没在 profile 里。

**教训**：硬件不同结果不同，实验耗时不能凭经验估，必须基准测试。

保留证据：`output/run2-v1-too-light-cpu/`

### Run 3 - 成功（慢路径加重 + 独立运行）

- 慢路径加重到 8 亿次乘法（~200ms，真实生产毛刺量级）
- Group A 独立运行（`run-experiment-part3.sh`），30s 采样
- Group B 独立运行（`run-experiment-part2.sh`），6 × 5s 采样
- 两次运行用相同的服务和压测配置

## 目录结构

```
e3-spike/
├── README.md                        ← 本文件
├── main.go                           ← HTTP 服务（带毛刺窗口逻辑）
├── run-experiment.sh                ← Part 1: 并发采样（失败，保留）
├── run-experiment-part2.sh          ← Part 2: 独立跑 Group B 时间序列
├── run-experiment-part3.sh          ← Part 3: 独立跑 Group A 30s 单窗口
└── output/
    ├── groupA-30s.pprof             ← Group A 最终成功数据
    ├── groupB-window-1.pprof ..     ← Group B 窗口 1-6 成功数据
    │   through groupB-window-6.pprof
    ├── groupB-timeseries/            ← Group B 副本
    ├── run1-v1-too-light-cpu/        ← Run 1 失败记录（教训证据）
    ├── run2-v1-too-light-cpu/        ← Run 2 失败记录（教训证据）
    ├── server-part3.log
    ├── hey-fast-part3.log / hey-slow-part3.log
    └── final-stats-part3.txt
```

## 核心数据对比

| 指标 | Group A 30s | Group B 窗口 4 | 差距 |
|------|:-----------:|:-------------:|:----:|
| `processSlowPath` flat% | 4.80% | **16.47%** | **3.43×** |
| 在 top 中的排名 | 第 5 | 第 2 | — |
| 能定位毛刺时段？ | ❌ | ✅ 15-20s | — |

详细分析：`../../output/e3-spike/e3-summary.md`

## 运行方式

```bash
# Part 2（先跑 Group B）
bash run-experiment-part2.sh

# Part 3（再跑 Group A）
bash run-experiment-part3.sh

# 分析
go tool pprof -top output/groupA-30s.pprof
for i in 1 2 3 4 5 6; do
  echo "=== 窗口 $i ==="
  go tool pprof -top output/groupB-window-$i.pprof | head -10
done
```

## 运行环境

- Go 1.26.2
- darwin/arm64（Apple Silicon）
- GOMAXPROCS=14
