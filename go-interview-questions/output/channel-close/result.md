# channel-close 运行结果

**对应题目**：八、关闭 channel
**目的**：确认关闭之后"读"与"写"的语义差异、重复关闭的后果、nil channel 的行为，以及接收方关闭时会发生什么
**运行方式**：`cd code/channel-close && go run .`
**原始输出**：`result.txt`（stdout）

## 关键输出

```
== 关闭之后继续接收 ==
  第 1 次接收：值=1 ok=true
  第 2 次接收：值=2 ok=true
  第 3 次接收：值=0 ok=false
  第 4 次接收：值=0 ok=false

== 关闭之后继续发送 ==
  向已关闭的 channel 发送
    -> panic: send on closed channel

== 重复关闭 ==
  同一个 channel 关闭两次
    -> panic: close of closed channel

== nil channel 上的收发都永久阻塞 ==
  在 nil channel 上接收（配 150ms 超时）
    -> 150ms 内没有任何进展，接收永久阻塞
  在 nil channel 上发送（配 150ms 超时）
    -> 150ms 内没有任何进展，发送永久阻塞

== 接收方关闭 channel，发送方会崩溃 ==
  接收方先 close，发送方随后发送
    -> 发送方 panic: send on closed channel
```

## 结论

关闭**只改"读"的语义**：缓冲里剩下的值可以正常取走，取完之后不再阻塞，直接返回零值和 `ok=false`。这是关闭唯一温和的表现。

"写"完全不受影响——关闭之后发送会 panic，重复关闭也会 panic。nil channel 的收发都永久阻塞（这个特性在 `select` 里有用：把某个 case 的 channel 设成 nil，该分支就不再被选中）。

最后一种最容易被忽略：接收方关闭 channel 会把发送方炸掉。这就是"只有发送方该关"这条规则的由来——关闭这个动作表达的是"我不会再发了"。
