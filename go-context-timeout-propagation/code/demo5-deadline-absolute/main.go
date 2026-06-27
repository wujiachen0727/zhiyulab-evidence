package main

import (
	"context"
	"fmt"
	"time"
)

// Demo 5: deadline 是绝对时间点，不是倒计时
// 证明：WithTimeout(ctx, 5s) 不是"从现在开始倒计时 5 秒"，
// 而是"设一个绝对时间点 = time.Now() + 5s"。
// 跨服务传播时，这意味着 deadline 会随传播路径衰减。

func main() {
	fmt.Println("=== Demo: deadline 是绝对时间点 ===")

	// 模拟 Service A 设置 5s timeout
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deadline, _ := ctx.Deadline()
	fmt.Printf("Service A 设置 timeout: 5s\n")
	fmt.Printf("Service A 当前时间: %s\n", now.Format("15:04:05.000"))
	fmt.Printf("deadline 绝对时间: %s\n\n", deadline.Format("15:04:05.000"))

	// 模拟网络传输耗时 1s
	fmt.Println("--- 网络传输耗时 1s ---")
	time.Sleep(1 * time.Second)

	// Service B 收到请求，查看 deadline
	remaining := time.Until(deadline)
	fmt.Printf("\nService B 收到请求时:\n")
	fmt.Printf("  当前时间: %s\n", time.Now().Format("15:04:05.000"))
	fmt.Printf("  deadline 仍然是: %s\n", deadline.Format("15:04:05.000"))
	fmt.Printf("  剩余时间: %v\n", remaining.Round(time.Millisecond))
	fmt.Println("  → deadline 没有'重置'，它在网络传输中衰减了 1s")

	// 再模拟 Service B 处理耗时 2s 后转发给 Service C
	fmt.Println("\n--- Service B 处理 2s 后转发给 Service C ---")
	time.Sleep(2 * time.Second)

	remaining = time.Until(deadline)
	fmt.Printf("\nService C 收到请求时:\n")
	fmt.Printf("  当前时间: %s\n", time.Now().Format("15:04:05.000"))
	fmt.Printf("  deadline 仍然是: %s\n", deadline.Format("15:04:05.000"))
	fmt.Printf("  剩余时间: %v\n", remaining.Round(time.Millisecond))
	fmt.Println("  → 从 A 设的 5s，经过传输+处理，C 只剩 ~2s")
	fmt.Println("\n结论: deadline 是绝对时间点。每经过一跳，可用时间就少一截。")
}
