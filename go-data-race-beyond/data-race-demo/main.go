// E1: 数据竞争 demo — 看起来安全的并发模式（v2）
// 场景：多个 goroutine 共享一个切片，各自往里面写
// 这是一个"看起来没问题"的代码——用了 append，觉得 append 是原子的
// 但实际上有 data race
//
// 运行: go run -race main.go
// Go 版本: 1.26.4 darwin/arm64

package main

import (
	"fmt"
	"sync"
)

func main() {
	// 常见场景：多个 worker 把结果收集到一个切片
	var results []int
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 看起来安全的操作——只是 append
			// 很多人以为 append 是原子的
			results = append(results, id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("results: %v (len=%d)\n", results, len(results))
}
