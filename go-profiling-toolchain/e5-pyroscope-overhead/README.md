# E5：Pyroscope 开销实测

## 实验目的

回应读者质疑："Grafana 官方声称 Pyroscope 持续 profiling 开销 <1%，真的吗？"——**自实测，给作者评价**。

## 核心结果

| 指标 | 基线 | 带 agent（全量 10 profile types）| 差距 |
|------|:----:|:-----------------------------:|:----:|
| QPS | 45388 | 44657 | **-1.61%** |
| 平均延迟 | 1.8 ms | 1.8 ms | 持平 |
| P95 延迟 | 0.7 ms | 0.8 ms | +14% |
| P99 延迟 | 0.9 ms | 0.8 ms | -11%（误差内） |
| 进程 CPU | 283.9% | 294.4% | **+3.70%** |
| 进程 RSS | 23.1 MB | 30.9 MB | **+7.8 MB（+33.8%）** |

**作者评价**（用于 R2 引用附加）：
> 实测比 Grafana 宣传的 `<1%` 略高，但我开了全量 10 种 profile types（比生产常见配置激进）。
> 对 99% 业务服务，用 **1-4% 性能换持续可观测**是划算的交易。

## 运行环境

- Go 1.26.2 / darwin arm64 / M 系列芯片 / GOMAXPROCS=14
- Pyroscope v1.21.0（Docker 容器 `grafana/pyroscope:latest`）
- Pyroscope Go SDK v1.2.8
- 压测工具：hey（`-c 30 -z 60s`）

## 目录结构

```
e5-pyroscope-overhead/
├── README.md                         ← 本文件
├── run-experiment.sh                 ← 实验脚本
├── baseline/
│   └── main.go                       ← 基线版（无 agent）
├── with-agent/
│   ├── main.go                       ← agent 版（Pyroscope Go SDK）
│   ├── go.mod / go.sum
└── output/
    ├── baseline-bin / agent-bin      ← 编译好的二进制
    ├── baseline-server.log / agent-server.log
    ├── baseline-hey.log / agent-hey.log    ← 压测详细输出
    ├── baseline-proc.csv / agent-proc.csv  ← 进程资源时间序列
    ├── baseline-final-stats.txt / agent-final-stats.txt
    ├── e5-comparison.txt             ← 核心对比报告
    └── run1-cpu-monitoring-broken/   ← Run 1 失败记录（go run 导致 PID 错）
```

## 运行方式

### 前置：启动 Pyroscope server
```bash
docker run -d --name pyroscope-e5 -p 4040:4040 grafana/pyroscope:latest
```

### 跑实验
```bash
bash run-experiment.sh
```

实验时长约 **3 分钟**：
- 60s 基线压测
- 60s 带 agent 压测
- 30s 服务启停和结果整理

### 查看结果
```bash
cat output/e5-comparison.txt
```

## 详细分析

见 `../../output/e5-pyroscope-overhead/e5-summary.md`。

## 意外发现

1. **P95/P99 无明显上升**——Pyroscope 的异步后台上传不阻塞请求路径
2. **RSS 增加比 CPU 开销更显著**（+33.8% vs +3.70%）——agent 需要维护 profile 缓冲区
3. **CPU profile 是全局单例**（来自 E3）—— 装了 Pyroscope agent 后，`go tool pprof http://.../debug/pprof/profile` 会偶发失败（agent 正在采样）。这是一个生产陷阱。
