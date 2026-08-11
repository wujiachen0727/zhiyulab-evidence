package main

import (
	"runtime"
	"testing"
)

var sink any

// 用 HeapObjects（对象数）对比 tiny 合并效果：
// N 个 4B 对象走 tiny allocator → 多个对象合并进 16B 槽位 → 对象数 < N
// N 个 16B 对象走 size class → 每个对象独立 → 对象数 = N
func TestTinyObjectCount(t *testing.T) {
	count := func(name string, n int, alloc func(i int) any) int64 {
		runtime.GC()
		var m0 runtime.MemStats
		runtime.ReadMemStats(&m0)
		base := m0.HeapObjects

		objs := make([]any, 0, n)
		for j := 0; j < n; j++ {
			objs = append(objs, alloc(j))
		}
		sink = objs

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		inc := int64(m1.HeapObjects) - int64(base)
		t.Logf("%s %d 个: 堆对象数增量 %d（合并比 %.1fx）", name, n, inc, float64(n)/float64(inc))
		return inc
	}

	count("4B 对象(tiny合并)", 1_000_000, func(i int) any { return &T4{} })
	count("8B 对象(tiny合并)", 1_000_000, func(i int) any { return &T8{} })
	count("16B 对象(sizeclass)", 1_000_000, func(i int) any { return &T16{} })
	count("32B 对象(sizeclass)", 1_000_000, func(i int) any { return &T32{} })
}
