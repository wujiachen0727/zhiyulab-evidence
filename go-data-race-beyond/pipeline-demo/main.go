// E5: 三 goroutine pipeline — 所有权模型实战
// 场景：生产者 → 处理器 → 消费者
// 每个阶段有明确的数据所有权
//
// 运行: go run -race main.go
// Go 版本: 1.26.4 darwin/arm64

package main

import (
	"fmt"
	"sync"
	"time"
)

// =============================================
// 无所有权版本：所有 goroutine 共享数据
// =============================================

func pipelineNoOwnership(items []int) {
	var processed []int
	var results []int
	var wg sync.WaitGroup

	// ���段 1：生成
	for _, item := range items {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			processed = append(processed, v*2) // 共享 processed — data race!
		}(item)
	}
	wg.Wait()

	// 阶段 2：处理
	for _, v := range processed {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			results = append(results, v+1) // 共享 results — data race!
		}(v)
	}
	wg.Wait()

	fmt.Printf("无所有权结果: %v\n", results)
}

// =============================================
// 有所有权版本：每个 goroutine 拥有自己的数据
// =============================================

// producer 拥有 raw 数据，产出发送到 ch1
func producer(items []int, ch1 chan<- int) {
	for _, item := range items {
		ch1 <- item
	}
	close(ch1)
}

// processor 拥有处理逻辑，从 ch1 接收，处理后发送到 ch2
func processor(ch1 <-chan int, ch2 chan<- int) {
	for v := range ch1 {
		ch2 <- v * 2
	}
	close(ch2)
}

// consumer 拥有最终结果，从 ch2 接收并收集
func consumer(ch2 <-chan int) []int {
	var results []int
	for v := range ch2 {
		results = append(results, v+1)
	}
	return results
}

func pipelineWithOwnership(items []int) {
	ch1 := make(chan int, 10)
	ch2 := make(chan int, 10)

	go producer(items, ch1)
	go processor(ch1, ch2)
	results := consumer(ch2)

	fmt.Printf("有所有权结果: %v (len=%d)\n", results, len(results))
}

func main() {
	items := []int{1, 2, 3, 4, 5}

	// 先跑无所有权的（可能有 data race，程序崩溃不影响）
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("无所有权版本 panic（data race 导致的）:", r)
			}
		}()
		pipelineNoOwnership(items)
	}()

	time.Sleep(10 * time.Millisecond)

	// 跑有所有权的
	pipelineWithOwnership(items)
}
