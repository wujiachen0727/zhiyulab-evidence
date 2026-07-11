# E1：发送端 msgID 超时补传

演示 at-least-once 投递的第一道防线——发送端用唯一 `msgID` 做超时补传。

## 运行
```bash
go run .
```
环境：Go 1.26.4 darwin/arm64，仅用标准库，无需外部依赖。

## 关键结论（见 output/sender-retry/result.txt）
- 第 1 次发送后 ACK 被网络丢弃，发送端超时。
- 发送端用**同一个 msgID** 重发（第 2 次），最终收到 ACK 投递成功。
- 核心：`msgID` 在重试中保持不变，重试 ≠ 新消息，接收端可据此去重。

## 与 IM 失败模式的绑定
对应"弱网丢包"场景：移动网络下 ACK 丢失是常态，at-least-once 用"重发 + 稳定 msgID"兜底，把"可能丢"变成"最终送达"。
