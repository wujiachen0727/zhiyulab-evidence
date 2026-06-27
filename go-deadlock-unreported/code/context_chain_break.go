// context 链断裂——goroutine 永久等待 ctx.Done()
// 运行：go run context_chain_break.go
//
// 典型场景：多层 goroutine 嵌套，父级监听 context 取消，
// 但子 goroutine 没有传递 context，或者没有 select ctx.Done()
//
// goroutine dump 特征：goroutine 卡在 chan receive（等待 ctx.Done()）
// 但因为 context 链已断裂，ctx.Done() 永远不会返回

package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// brokenWorker —— 没有监听 ctx.Done() 的子 goroutine
// 这是最常见的 context 链断裂模式
func brokenWorker(ctx context.Context, id int, results chan<- string) {
	for {
		// 模拟长时间工作
		time.Sleep(500 * time.Millisecond)

		select {
		case <-ctx.Done():
			fmt.Printf("[Worker%d] 收到取消信号，退出\n", id)
			return
		default:
			// 这里没有 select ctx.Done() —— 实际上 default 分支确保了不会阻塞
			// 真正的 context 链断裂场景是：
			// 子 goroutine 启动的 goroutine 没有传 ctx
		}
	}
}

// deepNestedGoroutine —— 深度嵌套的 context 链断裂
// 父 goroutine 收到了 ctx.Done()，但它启动的子 goroutine 没传 ctx
func deepNestedGoroutine(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// 子 goroutine 1：正确监听了 ctx
	go func() {
		<-ctx.Done()
		fmt.Println("[Sub1] 收到取消信号")
	}()

	// 子 goroutine 2：没有传 ctx！（使用 context.Background）
	// 这是最常见的 bug：在嵌套 goroutine 中用了 background context
	go func() {
		// 没有监听任何取消信号，永久等待
		<-context.Background().Done() // 永远等不到！
		fmt.Println("[Sub2] 这行永远不会执行")
	}()

	// 子 goroutine 3：有 context 但没有 select ctx.Done()
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			// 做了很多工作，但从不检查 ctx
			// 典型场景：for 循环中只做业务逻辑，没加 ctx.Done() 检查
		}
	}()

	<-ctx.Done()
	fmt.Println("[Parent] 收到取消信号")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deepNestedGoroutine(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	// 取消 context
	fmt.Println("[Main] 取消 context...")
	cancel()

	time.Sleep(300 * time.Millisecond)

	fmt.Println("\n=== goroutine dump 特征信号 ===")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	dump := string(buf[:n])

	if contains(dump, "chan receive") {
		fmt.Println("✅ 发现 chan receive 信号（goroutine 在等待 channel）")
	}
	if contains(dump, "Background()") {
		fmt.Println("✅ 发现 context.Background().Done()——这是 context 链断裂的典型信号")
	}

	fmt.Println("\n--- 关键 dump 片段（仅包含 context 相关 goroutine）---")
	lines := splitLines(dump)
	for _, line := range lines {
		if contains(line, "context") || contains(line, "chan receive") || contains(line, "Background") || contains(line, "Done") {
			fmt.Println(line)
		}
	}

	wg.Wait()
	fmt.Println("✅ 程序退出")
	fmt.Println("\n⚠️ 注意：Sub2 永远不会退出——因为它监听的是 context.Background().Done()")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
