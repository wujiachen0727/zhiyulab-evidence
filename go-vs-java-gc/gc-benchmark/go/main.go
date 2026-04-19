package main

import (
	"fmt"
	"runtime"
	"time"
)

// 全局变量防止编译器优化
var sink []*[64]byte

func main() {
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	gcBefore := m1.NumGC

	// 保持 10000 个存活对象（模拟工作集），不断替换制造 GC 压力
	liveSet := 10000
	sink = make([]*[64]byte, liveSet)
	allocs := 10_000_000

	start := time.Now()

	for i := 0; i < allocs; i++ {
		obj := new([64]byte)
		obj[0] = byte(i) // 防止优化
		sink[i%liveSet] = obj
	}

	elapsed := time.Since(start)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	gcCount := m2.NumGC - gcBefore
	totalPause := time.Duration(m2.PauseTotalNs - m1.PauseTotalNs)
	var avgPause time.Duration
	if gcCount > 0 {
		avgPause = totalPause / time.Duration(gcCount)
	}

	// 最大 STW：遍历最近的 pause 记录
	var maxPause time.Duration
	for i := uint32(0); i < gcCount && i < 256; i++ {
		idx := (m2.NumGC - 1 - i) % 256
		p := time.Duration(m2.PauseNs[idx])
		if p > maxPause {
			maxPause = p
		}
	}

	fmt.Printf("=== Go GC Benchmark ===\n")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Workload: %d allocs, live set %d × 64B\n", allocs, liveSet)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("GC cycles: %d\n", gcCount)
	fmt.Printf("GC CPU: %.2f%%\n", m2.GCCPUFraction*100)
	fmt.Printf("Total GC pause: %v\n", totalPause)
	fmt.Printf("Avg GC pause: %v\n", avgPause)
	fmt.Printf("Max GC pause: %v\n", maxPause)
	fmt.Printf("Heap alloc: %.2f MB\n", float64(m2.HeapAlloc)/1024/1024)
	fmt.Printf("Heap sys: %.2f MB\n", float64(m2.HeapSys)/1024/1024)
	fmt.Printf("Total alloc: %.2f MB\n", float64(m2.TotalAlloc)/1024/1024)
}
