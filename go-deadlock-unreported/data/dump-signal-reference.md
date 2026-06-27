# Dump 特征信号对照表

> 数据来源：实验运行结果（evidence/output/experiment-results.md）

## 三种阻塞模式的 goroutine dump 特征

| 阻塞模式 | dump 状态 | 关键函数 | 堆栈特征 | 是否报错 |
|---------|----------|---------|---------|:--------:|
| RWMutex 递归读阻塞 | `sync.RWMutex.RLock` | `runtime_SemacquireRWMutexR` | `GetData → RLock → sema` | ❌ |
| RWMutex writer 等待 | `sync.RWMutex.Lock` | `runtime_SemacquireRWMutex` | `RefreshCache → Lock → sema` | ❌ |
| Channel send 阻塞 | `chan send` | `runtime.chansend` | `ch <- msg → chansend` | ❌ |
| Channel receive 阻塞 | `chan receive` | `runtime.chanrecv` | `<-ch → chanrecv` | ❌ |
| Context 链断裂 | `chan receive (nil chan)` | `context.Background().Done` | 堆栈显示 context.Background | ❌ |
| 全局死锁（runtime 检测） | — | — | — | ✅ Go panic |

## 信号 → 模式反推矩阵

```
从 dump 信号到根因的快速定位：

看到大批 goroutine 在 sync.RWMutex.RLock
    → 判断：RWMutex 递归读阻塞
    → 找：有没有 goroutine 在 sync.RWMutex.Lock
    → 修复：避免在持有 RLock 时触发写锁操作

看到 goroutine 在 chan send
    → 判断：channel send 阻塞
    → 找：对应的接收方 goroutine 是否还在
    → 修复：errgroup + ctx.Done() 或 select 超时

看到 goroutine 在 chan receive (nil chan)
    → 判断：context 链断裂
    → 找：goroutine 创建时是否传了正确的 context
    → 修复：传递父 context，select 监听 ctx.Done()
```

## 三模式 × 信号 × 根因 × 修复

| 模式 | dump 信号 | 根因 | 修复方案 |
|:----:|:---------:|------|---------|
| RWMutex 递归读阻塞 | `sync.RWMutex.RLock` | writer-preference 设计 + 递归调用 | 命名约定 + 锁分解 |
| Channel send 阻塞 | `chan send` | 接收方退出后发送方继续发送 | errgroup + ctx.Done() |
| Context 链断裂 | `chan receive (nil chan)` | 子 goroutine 未监听 ctx.Done() | 传递 ctx + select 多路复用 |
