package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// chain 把同一个 ctx 沿调用链向下一层传，直到最底层才开始等取消
func chain(ctx context.Context, level, max int) {
	if level < max {
		chain(ctx, level+1, max)
		return
	}
	<-ctx.Done()
	fmt.Printf("  最底层（第 %d 层调用）收到取消，ctx.Err() = %v\n", level, ctx.Err())
}

func main() {
	fmt.Println("== 1. 取消怎么传到最底下 ==")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	chain(ctx, 1, 5)
	// 取整到 10ms：这个数字来自定时器（50ms）+ 调度抖动，精确到毫秒不可复现
	fmt.Printf("  从取消到最底层退出用了 %v\n", time.Since(start).Round(10*time.Millisecond))

	fmt.Println()
	fmt.Println("== 2. 不调用 cancel 的代价 ==")
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		child, _ := context.WithCancel(context.Background())
		go func() { <-child.Done() }()
	}
	time.Sleep(150 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  派生 100 个子 context 且不调用 cancel：协程数 %d -> %d\n", before, after)

	child, childCancel := context.WithCancel(context.Background())
	fmt.Printf("  未取消的子 context：Done() 是否已关闭 -> %v，Err() -> %v\n", isClosed(child), child.Err())
	childCancel()
	time.Sleep(20 * time.Millisecond)
	fmt.Printf("  调用 cancel() 之后：Done() 是否已关闭 -> %v，Err() -> %v\n", isClosed(child), child.Err())

	fmt.Println()
	fmt.Println("== 3. 超时和取消是同一个机制 ==")
	tctx, tcancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer tcancel()
	<-tctx.Done()
	fmt.Printf("  WithTimeout 到期后：Err() -> %v\n", tctx.Err())

	fmt.Println()
	fmt.Println("== 4. 值也沿同一条链往下走 ==")
	vctx := context.WithValue(context.Background(), traceKey{}, "req-7f3a")
	reportValue(vctx, 1, 3)
}

type traceKey struct{}

func reportValue(ctx context.Context, level, max int) {
	if level < max {
		reportValue(ctx, level+1, max)
		return
	}
	fmt.Printf("  第 %d 层拿到 trace id：%v\n", level, ctx.Value(traceKey{}))
}

func isClosed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
