# 实验输出

## E1: RWMutex 递归读阻塞

**运行方式**：`go run evidence/code/rwmutex_recursive_read.go`

### 关键输出
- ReaderA 持有读锁后，Writer 等待写锁（状态：`sync.RWMutex.Lock`）
- ReaderB 尝试 RLock → 被阻塞（状态：`sync.RWMutex.RLock`）
- Go runtime 不报错，但 goroutine dump 可见：
  - Writer: `goroutine 4 [sync.RWMutex.Lock]: sync.runtime_SemacquireRWMutex`
  - ReaderB: `goroutine 5 [sync.RWMutex.RLock]: sync.runtime_SemacquireRWMutexR`

### 特征信号
- `semacquire` — RWMutex 阻塞的统一入口
- `sync.RWMutex.Lock` — writer 在等锁
- `sync.RWMutex.RLock` — reader 被阻塞

## E2: RWMutex 复杂调用链

**运行方式**：`go run evidence/code/rwmutex_callchain.go`

### 关键输出
- 1 个 reader 持锁 + 1 个 writer 等待 + 5 个并发 reader 被阻塞
- dump 显示：
  - `goroutine 7 [sync.RWMutex.Lock]` — writer
  - `goroutine 33-37 [sync.RWMutex.RLock]` — 5 个 reader 全部卡住

### 特征信号
- 多个 goroutine 在 `sync.RWMutex.RLock` — 典型的大量读请求被阻塞场景

## E3: Channel send 阻塞（无缓冲）

**运行方式**：`go run evidence/code/chan_send_block_unbuffered.go`

### 关键输出
- Receiver 接收 2 条消息后退出
- Sender 发送第 3 条消息时永久阻塞
- dump 显示：`goroutine 3 [chan send]`

### 特征信号
- `chan send` — goroutine 在等待向 channel 发送
- 阻塞位置：`ch <- msg` 语句

## E4: Context 链断裂

**运行方式**：`go run evidence/code/context_chain_break.go`

### 关键输出
- Sub1（正确监听 ctx.Done()）→ 正常退出
- Sub2（监听 context.Background().Done()）→ 永久阻塞
- Sub3（不监听 ctx）→ 死循环
- dump 显示：`goroutine 8 [chan receive (nil chan)]`

### 特征信号
- `chan receive` — goroutine 在等待 channel
- `(nil chan)` — nil channel 永远不会返回
- `context.Background()` — 没有使用父 context
