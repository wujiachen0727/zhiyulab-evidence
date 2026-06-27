// E4: 碎片化场景实测 — 大量小对象分配/释放后观察 HeapSys vs HeapAlloc 差距
// 运行方式: go run e4_fragmentation.go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("初始 | HeapAlloc=%dKB | HeapSys=%dKB | 碎片率=%.1f%%\n",
		m.HeapAlloc/1024, m.HeapSys/1024, fragRate(&m))

	// 阶段1：大量小对象分配
	const totalObjects = 1000000
	const objectSize = 64 // 小对象
	pointers := make([]*[]byte, totalObjects)

	for i := 0; i < totalObjects; i++ {
		b := make([]byte, objectSize)
		pointers[i] = &b
	}

	runtime.ReadMemStats(&m)
	fmt.Printf("分配后 | HeapAlloc=%dKB | HeapSys=%dKB | 碎片率=%.1f%%\n",
		m.HeapAlloc/1024, m.HeapSys/1024, fragRate(&m))

	// 阶段2：隔一个释放一个（制造碎片化）
	for i := 0; i < totalObjects; i += 2 {
		pointers[i] = nil
	}
	runtime.GC()

	runtime.ReadMemStats(&m)
	fmt.Printf("隔一释放后 | HeapAlloc=%dKB | HeapSys=%dKB | 碎片率=%.1f%%\n",
		m.HeapAlloc/1024, m.HeapSys/1024, fragRate(&m))

	// 阶段3：分配更大的对象（无法复用释放的空间）
	for i := 0; i < 100000; i++ {
		_ = make([]byte, 4096) // 4KB 对象
	}
	runtime.GC()

	runtime.ReadMemStats(&m)
	fmt.Printf("大对象分配后 | HeapAlloc=%dKB | HeapSys=%dKB | 碎片率=%.1f%%\n",
		m.HeapAlloc/1024, m.HeapSys/1024, fragRate(&m))

	// 保持存活对象
	_ = pointers
}

func fragRate(m *runtime.MemStats) float64 {
	if m.HeapSys == 0 {
		return 0
	}
	return float64(m.HeapSys-m.HeapAlloc) / float64(m.HeapSys) * 100
}
