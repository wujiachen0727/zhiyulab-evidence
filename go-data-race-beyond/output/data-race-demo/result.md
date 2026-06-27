# E1 实验输出 — 多 goroutine 并发 append 到切片

**运行命令**：`go run -race main.go`
**Go 版本**：1.26.4 darwin/arm64
**运行时间**：2026-06-19

## 结果

```
Found 3 data race(s)
```

## 关键输出

```
WARNING: DATA RACE
Write at 0x00c0000d8000 by goroutine 8:
  main.main.func1()
      /Users/wujiachen/WriteCraft/articles/go-data-race-beyond/evidence/code/data-race-demo/main.go:27 +0xd4

Previous read at 0x00c0000d8000 by goroutine 7:
  main.main.func1()
      /Users/wujiachen/WriteCraft/articles/go-data-race-beyond/evidence/code/data-race-demo/main.go:27 +0x70
```

**最终输出**：`results: [0 3 7 6 4 9] (len=6)` — 丢了 4 个元素（本该有 10 个）

## 分析

- 10 个 goroutine 并发 `append` 到同一个切片
- append 不是原子的——它读 len、写 len、可能触发 growslice，这些步骤交织在一起
- 结果：数据丢失（len=6 而非 10），且有 data race
- 这是典型的"看起来安全的代码"——没写锁、没传指针、就是 append
- 但因为是共享切片，没有明确的"数据所有权"，就出问题了
