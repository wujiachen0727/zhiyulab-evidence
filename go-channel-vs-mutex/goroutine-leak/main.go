package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Channel 滥用场景：用 Channel 做请求-响应模式时，
// 如果消费者超时退出，而 resp channel 是 unbuffered 的，
// 生产者的发送操作永久阻塞 → goroutine 泄漏。

func leakyDemo() int {
	type request struct {
		data int
		resp chan int // unbuffered! 这是泄漏根源
	}

	ch := make(chan request, 100)

	// cache manager goroutine
	go func() {
		for req := range ch {
			// 模拟处理延迟
			time.Sleep(5 * time.Millisecond)
			req.resp <- req.data * 2 // 如果没人读 resp，永远阻塞在这里
		}
	}()

	var wg sync.WaitGroup

	// 50 个客户端，全部用极短超时
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp := make(chan int) // unbuffered — 泄漏根源
			ch <- request{data: id, resp: resp}

			select {
			case result := <-resp:
				_ = result
			case <-time.After(1 * time.Millisecond):
				// 超时放弃，但 cache manager 还在试图写 resp...
				return
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	return runtime.NumGoroutine()
}

func main() {
	fmt.Println("=== Channel 滥用导致 goroutine 泄漏演示 ===")
	before := runtime.NumGoroutine()
	fmt.Printf("启动前 goroutine 数: %d\n", before)

	after := leakyDemo()
	fmt.Printf("结束后 goroutine 数: %d\n", after)
	fmt.Printf("泄漏的 goroutine: %d\n", after-before)

	fmt.Println("\n根因: cache manager 阻塞在 req.resp <- ... 上，")
	fmt.Println("因为消费者已超时退出，没人读 unbuffered resp channel。")
	fmt.Println("\n修复: resp 用 buffered channel(cap=1)，写入不阻塞即使没人读。")
}
