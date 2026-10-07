# map-concurrent 运行结果

**对应题目**：六、map 并发写
**目的**：确认单协程写 nil map 与多协程并发写 map 的性质差异，以及 `recover` 对两者的效果
**运行方式**：`cd code/map-concurrent && go run .`
**原始输出**：`result.txt`（stdout）、`stderr.txt`（stderr）

## 关键输出

stdout：

```
panic 被 recover 接住 -> assignment to entry in nil map
```

stderr：

```
fatal error: concurrent map writes

goroutine 12 [running]:
internal/runtime/maps.fatal(...)
main.main.func2(0x7)
```

**退出码：2**

8 个协程各写 10 万次，每个协程内部都装了 `recover`，**一条都没有触发**。

## 结论

两种错误的性质不同。写 nil map 是运行时 panic，能被 `recover` 接住；并发写 map 是运行时致命错误——`fatal` 说明运行时把它归进致命错误那一类，直接终止进程。此时 `defer` 不会跑，日志不会落盘，优雅关闭不会执行。

所以这道题的第二层是：这类错误没有兜底方案，只能在设计上避免（加锁或改用并发安全的数据结构）。
