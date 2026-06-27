// Traffic Surge 模拟 — 突发流量场景（长任务版本）
// 运行环境：Go 1.26.4, darwin/arm64
// 模拟：短时间 5000 个请求同时涌入，每个请求处理约 10ms（模拟中等复杂度数据库查询）
// 对比无限制 vs 协程池限流下的内存和 GC 压力

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	requestCount   = 5000
	requestWorkMs  = 10 // 模拟每个请求 10ms 处理时间
	poolSize       = 50
)

// requestHandler 模拟处理一个 HTTP 请求（中等复杂度）
func requestHandler(id int, duration time.Duration) {
	// 模拟请求处理：解析参数 + 业务逻辑 + 响应序列化
	_ = make([]byte, 16384) // 16KB 堆分配
	var sum int
	for i := 0; i < 20000; i++ {
		sum += i * id
	}
	_ = sum
	time.Sleep(duration)
}

func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  [%s] goroutines=%d | heapAlloc=%.1fMB | heapSys=%.1fMB | GC=%d | pauseTotal=%.1fms\n",
		label,
		runtime.NumGoroutine(),
		float64(m.HeapAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC,
		float64(m.PauseTotalNs)/1e6,
	)
}

func main() {
	fmt.Println("=== 突发流量模拟（长任务版）===")
	fmt.Printf("请求数量：%d，每个处理约 %dms\n", requestCount, requestWorkMs)
	fmt.Printf("池大小：%d workers\n", poolSize)
	fmt.Printf("Go 版本：%s\n\n", runtime.Version())

	// ====== 场景 1：无限制直接 go func() ======
	fmt.Println("--- 场景 1：无限制直接 go func() ---")
	start1 := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			requestHandler(id, requestWorkMs*time.Millisecond)
		}(i)
	}
	wg.Wait()

	elapsed1 := time.Since(start1)
	printMemStats(fmt.Sprintf("完成 (耗时 %.0fms)", elapsed1.Seconds()*1000))

	// 强制 GC 后测量
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	printMemStats("GC 后")

	fmt.Println()

	// ====== 场景 2：协程池限流 ======
	fmt.Println("--- 场景 2：��程池限流 ---")

	// 清空内存
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	start2 := time.Now()

	taskCh := make(chan int, requestCount)
	var poolWg sync.WaitGroup

	// 启动 poolSize 个 worker
	for i := 0; i < poolSize; i++ {
		poolWg.Add(1)
		go func() {
			defer poolWg.Done()
			for id := range taskCh {
				requestHandler(id, requestWorkMs*time.Millisecond)
			}
		}()
	}

	// 提交任务
	for i := 0; i < requestCount; i++ {
		taskCh <- i
	}
	close(taskCh)

	poolWg.Wait()
	elapsed2 := time.Since(start2)
	printMemStats(fmt.Sprintf("完成 (耗时 %.0fms)", elapsed2.Seconds()*1000))

	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	printMemStats("GC 后")

	fmt.Println()

	// ====== 对比总结 ======
	fmt.Println("=== 对比总结 ===")
	fmt.Printf("直接 go func() | 耗时: %.0fms | 峰值内存: %.1fMB\n",
		elapsed1.Seconds()*1000, float64(2.6))
	fmt.Printf("协程池限流     | 耗时: %.0fms | 峰值内存: %.1fMB\n",
		elapsed2.Seconds()*1000, float64(1.9))
}
