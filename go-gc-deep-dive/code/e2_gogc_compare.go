// E2: GOGC 影响对比 — 足够重的负载让 GC 自然触发
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

func main() {
	gogc := debug.SetGCPercent(-1)
	debug.SetGCPercent(gogc)

	// 长期存活对象：10MB
	longLived := make([]*[]byte, 10000)
	for i := range longLived {
		b := make([]byte, 1024)
		longLived[i] = &b
	}

	// 大量短生命周期分配
	start := time.Now()
	for i := 0; i < 10000000; i++ {
		_ = make([]byte, 64)
	}
	elapsed := time.Since(start)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("GOGC=%d | NumGC=%d | PauseTotalMs=%d | HeapAllocMB=%d | HeapSysMB=%d | GCCPU%%=%.1f | Elapsed=%v\n",
		gogc, m.NumGC, m.PauseTotalNs/1000000, m.HeapAlloc/1024/1024, m.HeapSys/1024/1024,
		float64(m.GCCPUFraction)*100, elapsed)

	_ = longLived
}
