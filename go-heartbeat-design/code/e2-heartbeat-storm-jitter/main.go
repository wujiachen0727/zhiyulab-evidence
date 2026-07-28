package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	const numConnections = 10000
	const heartbeatInterval = 30 * time.Second

	// === 无 Jitter：所有连接同时发心跳 ===
	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)
	start0 := time.Now()

	var wg0 sync.WaitGroup
	for i := 0; i < numConnections; i++ {
		wg0.Add(1)
		go func() {
			defer wg0.Done()
			ticker := time.NewTicker(heartbeatInterval)
			defer ticker.Stop()
			select {
			case <-ticker.C:
			case <-time.After(100 * time.Millisecond): // 只等一次 tick
			}
		}()
	}
	wg0.Wait()
	elapsed0 := time.Since(start0)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	fmt.Printf("=== 心跳风暴实验 ===\n")
	fmt.Printf("连接数: %d, 心跳间隔: %v\n\n", numConnections, heartbeatInterval)

	fmt.Printf("--- 无 Jitter ---\n")
	fmt.Printf("一次 tick 耗时: %v\n", elapsed0)
	fmt.Printf("HeapInuse: %.2f MB\n", float64(m1.HeapInuse-m0.HeapInuse)/1024/1024)
	fmt.Printf("Goroutine 峰值: %d\n", runtime.NumGoroutine())

	// === 有 Jitter：每个连接增加随机偏移 ===
	runtime.GC()
	runtime.ReadMemStats(&m0)
	start1 := time.Now()

	var wg1 sync.WaitGroup
	for i := 0; i < numConnections; i++ {
		wg1.Add(1)
		go func(id int) {
			defer wg1.Done()
			// Jitter: 30s ± 5s
			jitter := time.Duration(id%10) * time.Second / 2
			interval := heartbeatInterval + jitter - 5*time.Second
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			select {
			case <-ticker.C:
			case <-time.After(200 * time.Millisecond):
			}
		}(i)
	}
	wg1.Wait()
	elapsed1 := time.Since(start1)

	runtime.ReadMemStats(&m1)

	fmt.Printf("\n--- 有 Jitter (30s ± 5s) ---\n")
	fmt.Printf("一次 tick 耗时: %v\n", elapsed1)
	fmt.Printf("HeapInuse: %.2f MB\n", float64(m1.HeapInuse-m0.HeapInuse)/1024/1024)
	fmt.Printf("Goroutine 峰值: %d\n", runtime.NumGoroutine())
}
