package main

import (
	"context"
	"fmt"
	"time"
)

// Demo 1: WithTimeout 的计时从调用瞬间开始，排队时间全算在内
// 证明：如果在 WithTimeout 之后有任何等待（排队、DNS、连接池等），
// 实际可用于"做事"的时间会比你以为的少。

func main() {
	// 你设了 3 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println("=== Demo: WithTimeout 计时起点 ===")
	fmt.Printf("设定超时: 3s\n")
	fmt.Printf("WithTimeout 调用时刻: %s\n", time.Now().Format("15:04:05.000"))

	// 模拟排队等待（比如等连接池、等 DNS 解析、等 TLS 握手）
	fmt.Println("\n模拟排队等待 2s...")
	time.Sleep(2 * time.Second)

	// 排队结束，现在才开始"真正做事"
	deadline, _ := ctx.Deadline()
	remaining := time.Until(deadline)
	fmt.Printf("\n排队结束，开始做事的时刻: %s\n", time.Now().Format("15:04:05.000"))
	fmt.Printf("剩余可用时间: %v\n", remaining.Round(time.Millisecond))
	fmt.Printf("你以为有 3s，实际只剩 %v\n", remaining.Round(time.Millisecond))

	// 尝试做一个 2 秒的操作——会超时
	fmt.Println("\n尝试执行一个需要 2s 的操作...")
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("操作完成 ✓")
	case <-ctx.Done():
		fmt.Printf("操作被取消 ✗ (原因: %v)\n", ctx.Err())
	}
}
