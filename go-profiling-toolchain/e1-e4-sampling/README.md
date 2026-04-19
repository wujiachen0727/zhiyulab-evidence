# E1 + E4：CPU 热点基线 + 采样频率对比

## 实验目的

- **E1**：证明 pprof 能清晰识别 CPU 热点
- **E4**：证明采样频率影响分辨率

## 🔥 意外之喜：两个反直觉发现

### 发现 1：pprof 的采样率硬编码 100Hz

查 Go 源码：`runtime/pprof/pprof.go` 里 `const hz = 100` —— **Go 把采样率写死在 pprof 包里**，你在外面 `runtime.SetCPUProfileRate(1000)` 根本不起作用。

Go 官方注释里还解释了原因：**操作系统信号频率上限约 500Hz**——Go 取了中间一个保守值。

### 发现 2：长函数在 profile 里被低估

实测 heavyFn 理论 24% → 实测 20.41%（**低估 15%**）。根因：Go 异步抢占每 10ms 打断一次长函数，这部分时间被归给 `runtime.asyncPreempt` 而不是 heavyFn。

## 实验数据核心对照

| 函数 | 单次耗时 | 理论 flat% | 实测 flat% | 结论 |
|------|:-------:|:---------:|:----------:|:----:|
| heavyFn | 120ms | ~24% | 20.41% | **被低估 15%**（异步抢占分摊）|
| mediumFn | 5ms | ~25% | 25.04% | 准确 |
| shortFn | 50μs | ~25% | 25.87% | 略高 |
| microFn | 5μs | ~25% | 25.30% | 略高 |

## 运行环境

- Go 1.26.2 / darwin arm64 / M 系列芯片 / GOMAXPROCS=14

## 运行方式

```bash
bash run-experiment.sh
```

会依次用 RATE_HZ=100/1000/10000 跑三次。

**注意**：三次实际采样率都是 100Hz，因为 pprof 硬编码——这正是本实验最重要的发现。

## 详细分析

见 `../../output/e1-e4-sampling/e1-e4-summary.md`。

## 源码证据

```go
// Go 1.26.2 runtime/pprof/pprof.go: StartCPUProfile 函数内
const hz = 100  // 硬编码

// Go 官方注释：
// "in practice operating systems cannot trigger signals at more than 
//  about 500 Hz, and our processing of the signal is not cheap"
```

## 目录结构

```
e1-e4-sampling/
├── README.md             ← 本文件
├── main.go               ← 工作负载（heavy/medium/short/micro 四种函数）
├── run-experiment.sh     ← 实验脚本
└── output/
    ├── cpu-100hz.pprof
    ├── cpu-1000hz.pprof  ← 内容和 100Hz 几乎相同（pprof 硬编码）
    ├── cpu-10000hz.pprof
    ├── run-100hz.log / run-1000hz.log / run-10000hz.log
    ├── e1-e4-comparison.txt
    └── workload          ← 编译后二进制
```
