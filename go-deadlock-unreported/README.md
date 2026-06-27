# 论据总索引

> 生成时间：2026-06-12
> 阶段：论证阶段

## 独立论据（自造）

| # | 类型 | 描述 | 文件位置 | 状态 |
|---|------|------|---------|:----:|
| E1 | 实验验证 | RWMutex 递归读阻塞最小复现代码 + dump 特征 | `code/rwmutex_recursive_read.go` | ✅ 运行通过 |
| E2 | 实验验证 | RWMutex 复杂调用链——多 reader + writer 阻塞 | `code/rwmutex_callchain.go` | ✅ 运行通过 |
| E3 | 实验验证 | Channel send 阻塞——接收方退出后发送方永久阻塞 | `code/chan_send_block_unbuffered.go` | ✅ 运行通过 |
| E4 | 实验验证 | Context 链断裂——子 goroutine 未监听 ctx.Done() | `code/context_chain_break.go` | ✅ 运行通过 |
| E5 | 场景模拟 | 线上排查路径——从告警到定位的完整流程 | `scenarios/debugging-scenario.md` | ✅ 完成 |
| E6 | 数据实测 | Dump 特征信号对照表——三模式 × 信号 × 根因 | `data/dump-signal-reference.md` | ✅ 完成 |

## 表达手法（不计入自造度）

| # | 类型 | 描述 | 说明 |
|---|------|------|------|
| E7 | 类比 | RWMutex 递归读阻塞 = "会议室的钥匙系统" | 辅助理解 |
| E8 | 类比 | Channel send 阻塞 = "传送带末端没人接货" | 辅助理解 |

## 外部引用

| # | 引用内容 | 来源 | 必要性 |
|---|---------|------|:------:|
| R1 | RWMutex writer 优先规则 | https://pkg.go.dev/sync#RWMutex | 关键前提佐证 |
| R2 | errgroup WithContext 取消传播 | https://pkg.go.dev/golang.org/x/sync/errgroup | 修复方案佐证 |
| R3 | Go runtime 死锁检测 checkdead | https://go.dev/src/runtime/proc.go | "为什么不报错"的佐证（可选） |

## 自造比例

- 独立论据合计：6 项（全部自造）
- 表达手法：2 项
- 外部引用：3 项（计划）
- **自造占比**：6 / (6+3) = **67%**
- **目标**：≥ 70%（需要在初稿中增加实测数据来提升）
