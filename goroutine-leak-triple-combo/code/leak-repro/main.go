// leak-repro: 最小复现 goroutine 泄漏
// 模拟三重组合：HTTP client 无超时 + 慢上游 + 无 context cancel
package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// slowServer 模拟一个响应极慢的上游服务
func startSlowServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟上游慢响应：每个请求阻塞 30 秒
		time.Sleep(30 * time.Second)
		w.Write([]byte("ok"))
	})
	go http.ListenAndServe(":18080", mux)
	time.Sleep(100 * time.Millisecond) // 等待 server 启动
}

func main() {
	startSlowServer()

	// 关键：使用默认 HTTP client（Timeout: 0 = 无超时）
	// 关键：没有 context cancel
	client := &http.Client{} // Timeout 默认为 0

	var wg sync.WaitGroup
	fmt.Println("开始发送请求...")
	fmt.Printf("初始 goroutine 数: %d\n", runtime.NumGoroutine())

	// 模拟高并发请求慢上游
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 没有 context.WithTimeout，没有 client.Timeout
			// 如果上游慢，这个 goroutine 会一直阻塞
			resp, err := client.Get("http://localhost:18080/slow")
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}(i)

		// 每 100 个请求打印一次 goroutine 数
		if (i+1)%100 == 0 {
			time.Sleep(50 * time.Millisecond)
			fmt.Printf("[%d 请求已发] goroutine 数: %d\n", i+1, runtime.NumGoroutine())
		}
	}

	// 持续监控 goroutine 数量
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("[等待 %ds] goroutine 数: %d\n", i+1, runtime.NumGoroutine())
	}

	fmt.Println("\n⚠️  1000 个 goroutine 全部阻塞在 http.Get 上，无人回收")
	fmt.Printf("最终 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println("在生产环境中，这个数字会持续增长直到 OOM")
}
