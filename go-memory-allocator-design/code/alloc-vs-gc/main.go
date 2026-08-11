// alloc-vs-gc — 分配成本 vs GC 成本分离实验（v4）
//
// 目标：验证"减少堆分配 = 减少 GC 压力——短生命周期对象是 GC 的主要负担"
// 方法：对比两种工作负载（对象都强制逃逸到堆）：
//   A. 每次分配新对象（短生命周期，分配后丢弃 → GC 回收）
//   C. 复用预分配对象池（零堆分配 → 无 GC 压力）
// 观察：耗时、GC 次数、堆对象增量
//
// 运行：go run .
package main

import (
	"fmt"
	"runtime"
	"time"
)

type point struct {
	x, y int
}

// go:noinline 强制对象逃逸到堆（返回值必须分配在堆上）
//
//go:noinline
func allocPoint(i int) *point {
	return &point{x: i, y: i}
}

// 对象池：预分配 1M 个对象，复用
var pool []*point

func init() {
	pool = make([]*point, 0, 1_000_000)
	for i := 0; i < 1_000_000; i++ {
		pool = append(pool, &point{})
	}
	// 强制逃逸（预分配也走堆）
	runtime.KeepAlive(pool)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const n = 10_000_000

	// 预热
	for i := 0; i < 100_000; i++ {
		p := allocPoint(i)
		runtime.KeepAlive(p)
	}

	// A: 每次分配新对象（短生命周期 → GC 回收）
	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)
	start := time.Now()
	for i := 0; i < n; i++ {
		p := allocPoint(i)
		runtime.KeepAlive(p)
	}
	dA := time.Since(start)
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// C: 复用对象池（零分配 → 无 GC 压力）
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	start = time.Now()
	idx := 0
	for i := 0; i < n; i++ {
		p := pool[idx%1_000_000]
		p.x, p.y = i, i
		runtime.KeepAlive(p)
		idx++
	}
	dC := time.Since(start)
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	fmt.Println("\n=== 分配 vs 复用（n =", n, "）===")
	fmt.Printf("A 每次分配新对象: %v | GC 次数 +%d | 堆对象增量 %d\n", dA, m1.NumGC-m0.NumGC, m1.HeapObjects-m0.HeapObjects)
	fmt.Printf("C 复用对象池:     %v | GC 次数 +%d | 堆对象增量 %d\n", dC, m3.NumGC-m2.NumGC, m3.HeapObjects-m2.HeapObjects)
	fmt.Printf("\nA/C = %.1fx —— 短生命周期分配触发 GC，是复用方案的 %d 倍耗时\n", float64(dA)/float64(dC), int(float64(dA)/float64(dC)+0.5))
	fmt.Printf("A 触发 GC %d 次（对象垃圾反复产生），C 仅 %d 次（零分配无垃圾）\n", m1.NumGC-m0.NumGC, m3.NumGC-m2.NumGC)
}
