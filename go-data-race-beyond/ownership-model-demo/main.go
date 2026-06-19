// E3: 数据所有权模型 demo
// 展示：同一需求，两种实现——"无所有权" vs "有所有权"
//
// 需求：多个 worker 并发处理任务，收集处理结果
// 
// 场景 A（无所有权）：所有 goroutine 共享一个 results 切片
// 场景 B（有所有权）：一个 goroutine 拥有 results，其他通过 channel 发送
//
// 运行: go run -race main.go
// Go 版本: 1.26.4 darwin/arm64

package main

import (
	"fmt"
	"sync"
)

// =============================================
// 场景 A：无所有权模型
// =============================================

func scenarioNoOwnership(tasks []int) []int {
	var results []int
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		go func(task int) {
			defer wg.Done()
			// 所有 goroutine 都往同一个 results 里写
			// 没有明确的"谁拥有 results"
			results = append(results, task*2)
		}(t)
	}

	wg.Wait()
	return results
}

// =============================================
// 场景 B：有所有权模型
// =============================================

func scenarioWithOwnership(tasks []int) []int {
	results := make([]int, 0, len(tasks))
	resultCh := make(chan int, len(tasks))
	var wg sync.WaitGroup

	// Worker goroutines: 处理任务，结果发到 channel
	for _, t := range tasks {
		wg.Add(1)
		go func(task int) {
			defer wg.Done()
			resultCh <- task * 2
		}(t)
	}

	// 等待所有 worker 完成，然后关闭 channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 主 goroutine 拥有 results 的所有权
	// 只有它往 results 里写
	for r := range resultCh {
		results = append(results, r)
	}

	return results
}

func main() {
	tasks := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	fmt.Println("=== 场景 A：无所有权模型 ===")
	r1 := scenarioNoOwnership(tasks)
	fmt.Printf("结果: %v (len=%d)\n", r1, len(r1))

	fmt.Println("\n=== 场景 B：有所有权模型 ===")
	r2 := scenarioWithOwnership(tasks)
	fmt.Printf("结果: %v (len=%d)\n", r2, len(r2))
}
