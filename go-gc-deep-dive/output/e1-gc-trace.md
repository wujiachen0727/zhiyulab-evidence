# E1：GC Trace 实测数据

> 🟢已验证 — 本机实测，Go 1.26.2 darwin/arm64, Apple M3 Max, 14 P

## 运行命令

```bash
GODEBUG=gctrace=1 go run e1_gc_trace.go
```

## 实测 GC Trace 输出（典型样本）

```
gc 1  @0.039s 0%: 0.092+0.85+0.033 ms clock, 1.2+0.49/1.3/0+0.46 ms cpu, 3->5->0 MB, 4 MB goal, 14 P
gc 10 @0.062s 2%: 0.071+0.32+0.005 ms clock, 0.99+0.27/0.51/0.027+0.072 ms cpu, 3->3->1 MB, 4 MB goal, 14 P
gc 20 @0.080s 2%: 0.052+0.19+0.003 ms clock, 0.73+0.090/0.47/0.13+0.052 ms cpu, 3->3->1 MB, 4 MB goal, 14 P
gc 30 @0.498s 0%: 0.37+1.5+0.75 ms clock, 5.2+0.17/4.1/0.031+10 ms cpu, 4->5->2 MB, 5 MB goal, 14 P
```

## 数据解读

| 指标 | 值 | 说明 |
|------|---|------|
| STW 初始暂停 | 0.025-0.37ms | 即 gc trace 输出的第一个数字（SweepTermination+MarkPhase） |
| 并发标记耗时 | 0.18-2.8ms | 第二个数字，与 mutator 并发运行 |
| STW 结束暂停 | 0.002-0.75ms | 第三个数字（MarkTermination） |
| GC CPU 占比 | 0-2% | 稳态运行时 GC CPU 开销极低 |

**关键结论**：现代 Go（1.26）的 STW 暂停确实在亚毫秒级别，与"从 300ms 到亚毫秒"的叙事一致。
