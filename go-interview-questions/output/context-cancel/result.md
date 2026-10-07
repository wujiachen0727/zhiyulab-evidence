# context-cancel 运行结果

**对应题目**：九、`context`
**目的**：确认取消如何沿调用链传播、不调用 `cancel` 的代价、超时与取消的关系
**运行方式**：`cd code/context-cancel && go run .`
**原始输出**：`result.txt`（stdout）

## 关键输出

```
最底层（第 5 层调用）收到取消，ctx.Err() = context canceled
从取消到最底层退出用了 50ms

派生 100 个子 context 且不调用 cancel：协程数 1 -> 101
未取消的子 context：Done() 是否已关闭 -> false，Err() -> <nil>
调用 cancel() 之后：Done() 是否已关闭 -> true，Err() -> context canceled

WithTimeout 到期后：Err() -> context deadline exceeded

第 3 层拿到 trace id：req-7f3a
```

## 结论

一次 `cancel()` 调用穿过五层返回，最底层的阻塞立刻解开。它靠的是 `Done()` 这个通道的关闭——通道关闭会唤醒所有等它的协程，取消信号就是这么广播下去的。

忘记 `cancel` 的代价很直接：子 context 一直活着，等它的协程也一直活着，协程数单调上涨。这就是第七节那个泄漏形态的来源。

`WithTimeout` 和 `cancel` 走的是同一套机制，区别只是谁按下了按钮。所以拿到 `cancel` 函数就要 `defer cancel()`，哪怕已经设了超时。

**计时说明**：50ms 是取整到 10ms 之后的值（原值来自 50ms 定时器加调度抖动，精确到毫秒不可复现）。
