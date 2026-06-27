// E3: GOMEMLIMIT + GOGC 组合效果
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
	limit := debug.SetMemoryLimit(-1)
	debug.SetMemoryLimit(limit)

	// 逐步增长堆
	objects := make([]*[]byte, 0, 20000)

	start := time.Now()
	for i := 0; i < 20000; i++ {
		b := make([]byte, 1024)
		objects = append(objects, &b)

		if i%2000 == 0 && i > 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("i=%d | HeapAllocMB=%d | HeapSysMB=%d | NumGC=%d\n",
				i, m.HeapAlloc/1024/1024, m.HeapSys/1024/1024, m.NumGC)
		}
	}
	elapsed := time.Since(start)

	// 释放一半
	half := len(objects) / 2
	for i := 0; i < half; i++ {
		objects[i] = nil
	}
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("释放后 | HeapAllocMB=%d | HeapSysMB=%d | NumGC=%d | GOGC=%d | GOMEMLIMIT=%dMB | Elapsed=%v\n",
		m.HeapAlloc/1024/1024, m.HeapSys/1024/1024, m.NumGC, gogc, limit/1024/1024, elapsed)

	_ = objects[half:]
}
