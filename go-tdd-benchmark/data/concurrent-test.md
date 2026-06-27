# E7: 并发测试确定性对比

[推演] 2026-04-14

## 降级说明

原计划：用同一套并发代码，分别用 Go 1.22 前后版本测试，展示循环变量陷阱和 synctest 的差异。

降级原因：当前环境仅 Go 1.26.2，无法安装多版本 Go。降级为逻辑推演+官方文档引用。

## Go 1.22 之前的并发测试痛点

### 1. 循环变量陷阱

Go 1.22 之前，for 循环变量在每次迭代中复用同一地址：

```go
// Go 1.22 前：这段代码有 bug
for _, tt := range tests {
    tt := tt  // 必须！否则 t.Parallel 里的 tt 是最后一个值
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // 使用 tt...
    })
}
```

这是 Go 测试中最经典的坑之一。Dave Cheney 专门讨论过这个"动态作用域"问题。

### 2. 并发确定性

Go 1.24 之前，没有官方机制让并发代码的测试变得确定性：

```go
// 这段测试要么 flaky 要么慢
func TestConcurrentSafe(t *testing.T) {
    var wg sync.WaitGroup
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ch <- i  // 调度不确定
        }()
    }
    wg.Wait()
    close(ch)
    // 断言 channel 中的值... 顺序不确定！
}
```

**"We can make the test less flaky at the expense of making it slower, and we can make it less slow at the expense of making it flakier, but we can't make it both fast and reliable."**
——Go 官方博客

### 3. synctest（Go 1.24+）

synctest 包提供了确定性并发测试：

```go
func TestConcurrentWithSynctest(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        // 在这个作用域内，goroutine 调度是确定性的
        ch := make(chan int)
        go func() { ch <- 42 }()
        // 不需要 time.Sleep，synctest 保证 goroutine 已执行
        got := <-ch
        if got != 42 {
            t.Errorf("got %d, want 42", got)
        }
    })
}
```

## 推演结论

1. **Go 存在了 15 年没有官方并发测试方案**（2009-2024）——这是 TDD 最致命的盲区：Go 的杀手级特性（goroutine/channel）恰好是 TDD 最不擅长的领域。
2. **synctest 的出现是对"并发测试长期缺失"的承认**——但 15 年的空白期意味着大量 Go 项目在缺乏确定性并发测试的情况下运行。
3. **这恰恰说明 Go 的测试哲学在进化，而非停滞**——Go 不是拒绝改进测试体系，而是选择在正确的时机引入正确的方案（而非匆忙推出半成品）。
