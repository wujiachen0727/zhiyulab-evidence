package main

import (
	"context"
	"fmt"
	"time"
)

// Demo 2: 子 context 设更长的 timeout 无效——最短者胜
// 证明：父 context 3s，子 context 10s，子在第 4 秒就被取消了。
// 因为父取消时，所有子必然取消。

func main() {
	fmt.Println("=== Demo: 子 context 不能比父长 ===")

	// 父 context: 3 秒
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer parentCancel()

	// 子 context: 10 秒（你以为它能活 10 秒）
	childCtx, childCancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer childCancel()

	parentDeadline, _ := parentCtx.Deadline()
	childDeadline, _ := childCtx.Deadline()

	fmt.Printf("父 context deadline: %s (3s后)\n", parentDeadline.Format("15:04:05.000"))
	fmt.Printf("子 context deadline: %s (你以为是10s后)\n", childDeadline.Format("15:04:05.000"))
	fmt.Printf("实际子 deadline == 父 deadline? %v\n\n", childDeadline.Equal(parentDeadline))

	// 等待子 context 被取消
	start := time.Now()
	<-childCtx.Done()
	elapsed := time.Since(start)

	fmt.Printf("子 context 存活时间: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("子 context 取消原因: %v\n", childCtx.Err())
	fmt.Println("\n结论: 子 context 设 10s 没用，父 3s 到期时子一起死。")
}
