// E4: channel 传切片的所有权转移
// 展示：通过 channel 传递切片时，发送方和接收方的所有权边界
//
// 场景：一个 goroutine 生成数据，通过 channel 传给另一个 goroutine 处理
// 关键问题：发送后还能不能碰这个切片？
//
// 运行: go run -race main.go
// Go 版本: 1.26.4 darwin/arm64

package main

import (
	"fmt"
	"sync"
)

// =============================================
// 错误模式：发送后继续使用切片（所有权不清）
// =============================================

func wrongTransfer() {
	data := make([]int, 100)
	ch := make(chan []int)

	go func() {
		// 接收方拿到 data
		received := <-ch
		received[0] = 999 // 修改切片内容
	}()

	// 发送 data
	ch <- data
	// 发送后继续使用 data — 问题就在这里
	// data 的"所有权"没有明确转移
	for i := 0; i < 100; i++ {
		data[i] = i
	}
}

// =============================================
// 正确模式：发送后不再碰切片（所有权转移）
// =============================================

func correctTransfer() {
	ch := make(chan []int)
	var wg sync.WaitGroup

	// 生产者：生成数据，发送后不再碰
	wg.Add(1)
	go func() {
		defer wg.Done()
		data := make([]int, 100)
		for i := 0; i < 100; i++ {
			data[i] = i
		}
		// 发送 data — 所有权转移给接收方
		ch <- data
		// 发送后不再碰 data
	}()

	// 消费者：接收数据，拥有完全所有权
	wg.Add(1)
	go func() {
		defer wg.Done()
		received := <-ch
		// 现在 received 归这个 goroutine 所有
		received[0] = 999 // 安全
		_ = fmt.Sprintf("received[0]=%d", received[0])
	}()

	wg.Wait()
}

func main() {
	fmt.Println("=== 错误模式：发送后继续使用 ===")
	wrongTransfer()

	fmt.Println("\n=== 正确模式：所有权转移 ===")
	correctTransfer()
}
