package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
)

// E1 补充实验：sync.Map vs bigcache 的 GC 压力对比
// 核心假设：sync.Map 存 interface{} 导致 GC 需要扫描所有指针，
// bigcache 用连续字节数组避免 GC 扫描。
// 在 key 数量大时（100万），GC 暂停差异明显。

const (
	numKeys   = 1_000_000
	valSize   = 64
)

func main() {
	fmt.Println("=== GC 压力对比实验 ===")
	fmt.Printf("Key 数量: %d, Value 大小: %d bytes\n\n", numKeys, valSize)

	// 测试 sync.Map
	fmt.Println("--- sync.Map ---")
	testSyncMapGC()

	// 强制 GC 清理
	runtime.GC()
	time.Sleep(time.Second)

	// 测试 bigcache
	fmt.Println("\n--- bigcache ---")
	testBigcacheGC()
}

func testSyncMapGC() {
	var m sync.Map
	val := make([]byte, valSize)
	rand.Read(val)

	// 填充 100万 key
	start := time.Now()
	for i := 0; i < numKeys; i++ {
		v := make([]byte, valSize)
		copy(v, val)
		m.Store(strconv.Itoa(i), v)
	}
	fmt.Printf("填充耗时: %v\n", time.Since(start))

	// 获取内存统计
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("HeapAlloc: %d MB\n", ms.HeapAlloc/1024/1024)
	fmt.Printf("HeapObjects: %d\n", ms.HeapObjects)

	// 手动触发 GC 并测量暂停时间
	gcPauses := measureGCPauses(5)
	fmt.Printf("GC 暂停时间（5次均值）: %v\n", gcPauses)
}

func testBigcacheGC() {
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.Verbose = false
	cache, _ := bigcache.New(context.Background(), config)
	val := make([]byte, valSize)
	rand.Read(val)

	// 填充 100万 key
	start := time.Now()
	for i := 0; i < numKeys; i++ {
		cache.Set(strconv.Itoa(i), val)
	}
	fmt.Printf("填充耗时: %v\n", time.Since(start))

	// 获取内存统计
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("HeapAlloc: %d MB\n", ms.HeapAlloc/1024/1024)
	fmt.Printf("HeapObjects: %d\n", ms.HeapObjects)

	// 手动触发 GC 并测量暂停时间
	gcPauses := measureGCPauses(5)
	fmt.Printf("GC 暂停时间（5次均值）: %v\n", gcPauses)

	cache.Close()
}

func measureGCPauses(n int) time.Duration {
	var totalPause time.Duration
	for i := 0; i < n; i++ {
		var ms1, ms2 runtime.MemStats
		runtime.ReadMemStats(&ms1)
		runtime.GC()
		runtime.ReadMemStats(&ms2)
		pause := ms2.PauseTotalNs - ms1.PauseTotalNs
		totalPause += time.Duration(pause)
	}
	return totalPause / time.Duration(n)
}
