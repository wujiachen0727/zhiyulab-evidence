// Controlled OOM Demo — 受控版本，创建有限数量 goroutine 观察内存增长趋势
// 运行环境：Go 1.26.4, darwin/arm64
// 本程序创建指定数量的 goroutine（默认 10 万），采集内存数据后安全退出

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	totalGoroutines = 100000
	batchSize       = 1000
)

func main() {
	fmt.Println("=== 受控 OOM 实验 ===")
	fmt.Printf("目标 goroutine 数量：%d\n", totalGoroutines)
	fmt.Printf("Go 版本：%s\n", runtime.Version())
	fmt.Printf("CPU：%d 核\n\n", runtime.NumCPU())

	// 打印初始状态
	printMemStats("初始")

	var wg sync.WaitGroup

	for i := 0; i < totalGoroutines/batchSize; i++ {
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// 每个 goroutine 分配 4KB 堆内存
				_ = make([]byte, 4096)
				time.Sleep(5 * time.Second)
			}(i*batchSize + j)
		}

		// 每批打印一次内存状态
		created := (i + 1) * batchSize
		printMemStats(fmt.Sprintf("已创建 %d", created))

		// 让调度器有机会运行
		runtime.Gosched()
	}

	fmt.Println("\n=== 所有 goroutine 已创建，等待完成 ===")
	wg.Wait()

	fmt.Println("\n=== 实验结束 ===")
	printMemStats("结束")
}

func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("[%s] goroutines=%d | heapAlloc=%.2f MB | heapSys=%.2f MB | heapObjects=%d | GC次数=%d\n",
		label,
		runtime.NumGoroutine(),
		float64(m.HeapAlloc)/1024/1024,
		float64(m.HeapSys)/1024/1024,
		m.HeapObjects,
		m.NumGC,
	)
}
