# 逻辑推演：语言能力边界不等于默认心智模型

## 前提

1. 一门语言通常能支持多种并发写法。
2. 但语言、标准库和社区惯例会让某些写法更顺手、更常见。
3. 文章讨论的是“默认表达成本”，不是“能力上限”。

## 推演

- Go 可以用 mutex、atomic、WaitGroup；所以不能写成“Go 只能 CSP”。但 goroutine + channel + context 让协作关系更容易被写到代码结构里。
- Java 可以写 Actor，也可以用 CompletableFuture；所以不能写成“Java 只能传统线程”。但 virtual thread 的核心价值是保留 thread-per-request 的同步写法，并降低大量阻塞等待的资源成本。
- Erlang 不只是 mailbox；完整工程还涉及 OTP supervision、link、monitor、分布式语义。把 Erlang 简化成“消息队列”会漏掉它最强的失败处理心智。

## 正文可用结论

语言能力像工具箱，默认心智模型像你最顺手拿起来的那把工具。本文比较的是后者：当你面对同一个任务编排问题时，哪种责任分配最自然地浮到代码表面。
