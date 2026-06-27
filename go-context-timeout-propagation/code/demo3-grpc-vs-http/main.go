package main

import (
	"context"
	"fmt"
	"time"
)

// Demo 3: gRPC 自动传播 deadline，HTTP 不会
// 这个 demo 用伪代码+输出对比展示差异，因为完整的 gRPC 服务需要 proto 文件。
// 核心证明：gRPC 通过 grpc-timeout metadata header 自动传播 deadline，
// 而 HTTP 调用需要手动编码 deadline 到 header 中。

func main() {
	fmt.Println("=== Demo: gRPC vs HTTP 超时传播 ===")
	fmt.Println()

	// 模拟上游设置 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deadline, _ := ctx.Deadline()
	fmt.Printf("上游设置 timeout: 5s (deadline: %s)\n\n", deadline.Format("15:04:05.000"))

	// === gRPC 场景 ===
	fmt.Println("--- 场景 1: gRPC 调用 ---")
	fmt.Println("上游调用 grpc.Invoke(ctx, ...) 时:")
	fmt.Println("  → gRPC 框架自动将 deadline 编码为 grpc-timeout header")
	fmt.Println("  → 下游 gRPC handler 收到请求时:")
	fmt.Println("    ctx 已经带有 deadline (自动重建)")
	fmt.Println("    deadline, ok := ctx.Deadline() // ok=true, deadline=上游传来的时间点")
	fmt.Println()

	// === HTTP 场景 ===
	fmt.Println("--- 场景 2: HTTP 调用 ---")
	fmt.Println("上游调用 http.Client.Do(req.WithContext(ctx)) 时:")
	fmt.Println("  → HTTP 协议没有 deadline 传播机制")
	fmt.Println("  → client 端 context 控制的是'本次 HTTP 请求的超时'")
	fmt.Println("  → 下游 HTTP handler 收到请求时:")
	fmt.Println("    ctx = request.Context() // 这是下游自己的 context，没有上游 deadline")
	fmt.Println("    deadline, ok := ctx.Deadline() // ok=false !!!")
	fmt.Println()

	// === 对比结论 ===
	fmt.Println("=== 对比结论 ===")
	fmt.Println("┌──────────────┬────────────────────────┬─────────────────────────┐")
	fmt.Println("│              │ gRPC                   │ HTTP                    │")
	fmt.Println("├──────────────┼────────────────────────┼─────────────────────────┤")
	fmt.Println("│ 传播机制     │ grpc-timeout header    │ 无原生机制              │")
	fmt.Println("│              │ (框架自动)             │ (需手动)                │")
	fmt.Println("├──────────────┼────────────────────────┼─────────────────────────┤")
	fmt.Println("│ 下游 ctx     │ 带 deadline            │ 无 deadline             │")
	fmt.Println("├──────────────┼────────────────────────┼─────────────────────────┤")
	fmt.Println("│ 超时衰减     │ 自动（传输耗时扣除）   │ 不衰减（因为根本没传）  │")
	fmt.Println("├──────────────┼────────────────────────┼─────────────────────────┤")
	fmt.Println("│ 混合架构风险 │ 低                     │ 高（容易遗漏）          │")
	fmt.Println("└──────────────┴────────────────────────┴─────────────────────────┘")
	fmt.Println()
	fmt.Println("结论: 在 gRPC + HTTP 混合架构中，gRPC 链路自动传播 deadline，")
	fmt.Println("     HTTP 链路默默丢掉。如果你只在入口设了 timeout，")
	fmt.Println("     HTTP 下游服务根本不知道上游还有多少时间。")

	// === 正确姿势 ===
	fmt.Println("\n=== HTTP 手动传播的正确姿势 ===")
	fmt.Println(`
// 发送端: 将 deadline 编码到 header
deadline, ok := ctx.Deadline()
if ok {
    remaining := time.Until(deadline)
    req.Header.Set("X-Request-Timeout-Ms",
        strconv.FormatInt(remaining.Milliseconds(), 10))
}

// 接收端: 从 header 重建 context
timeoutMs := req.Header.Get("X-Request-Timeout-Ms")
if timeoutMs != "" {
    ms, _ := strconv.ParseInt(timeoutMs, 10, 64)
    ctx, cancel = context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
    defer cancel()
}`)
}
