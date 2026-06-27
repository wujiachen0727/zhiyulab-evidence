// OOM Demo — 演示无限制创建 goroutine 导致内存暴涨
// 运行环境：Go 1.26.4, darwin/arm64
//
// 注意：本程序会持续消耗内存直至 OOM，建议在资源受限环境下运行
// 或在启动后手动 Ctrl+C 终止观察趋势

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)

var goroutineCount atomic.Int64

func main() {
	// 捕获 Ctrl+C 安全退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n=== 收到退出信号，打印最终状态 ===")
		printMemStats()
		os.Exit(0)
	}()

	// 定时打印内存和 goroutine 统计
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			printMemStats()
			count := goroutineCount.Load()
			if count > 0 && count%50000 == 0 {
				fmt.Printf("  [标记点] goroutine 数量已达 %d\n", count)
			}
		}
	}()

	fmt.Println("=== OOM Demo 开始 ===")
	fmt.Println("创建 goroutine 中... 观察内存增长趋势")
	fmt.Println("按 Ctrl+C 安全退出\n")

	var i int64
	for i = 1; ; i++ {
		go leakyGoroutine(i)
		goroutineCount.Store(i)

		// 每 10000 个 goroutine 稍作暂停让 GC 有机会运行
		if i%10000 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// leakyGoroutine 模拟一个有内存分配的 goroutine
func leakyGoroutine(id int64) {
	// 每个 goroutine 分配一块内存并持有引用
	_ = make([]byte, 4096) // 4KB 堆分配
	time.Sleep(10 * time.Second)
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("[%s] goroutines=%d | heapAlloc=%.2f MB | sys=%.2f MB | GC=%d\n",
		time.Now().Format("15:04:05"),
		runtime.NumGoroutine(),
		float64(m.HeapAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC,
	)
}
