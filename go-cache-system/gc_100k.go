package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
	"os"
	"bytes"

	"github.com/allegro/bigcache/v3"
)

// 补测：10万 key 场景下的 HeapObjects + GC CPU 占比
// 目的：拐点一的阈值声明需要同一规模的数据支撑

func main() {
	fmt.Println("=== 10万 key GC 压力对比 ===")
	fmt.Println()

	test10k_syncMap()
	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	test10k_bigcache()

	fmt.Println("\n=== 100万 key GC CPU 占比实测 ===")
	fmt.Println()
	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	testGCCPU_syncMap()
}

func test10k_syncMap() {
	var m sync.Map
	val := make([]byte, 64)
	rand.Read(val)

	for i := 0; i < 100_000; i++ {
		v := make([]byte, 64)
		copy(v, val)
		m.Store(strconv.Itoa(i), v)
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("sync.Map (10万 key):\n")
	fmt.Printf("  HeapObjects: %d\n", ms.HeapObjects)
	fmt.Printf("  HeapAlloc: %d MB\n", ms.HeapAlloc/1024/1024)
}

func test10k_bigcache() {
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.Verbose = false
	cache, _ := bigcache.New(context.Background(), config)
	val := make([]byte, 64)
	rand.Read(val)

	for i := 0; i < 100_000; i++ {
		cache.Set(strconv.Itoa(i), val)
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("bigcache (10万 key):\n")
	fmt.Printf("  HeapObjects: %d\n", ms.HeapObjects)
	fmt.Printf("  HeapAlloc: %d MB\n", ms.HeapAlloc/1024/1024)
	cache.Close()
}

func testGCCPU_syncMap() {
	var m sync.Map
	val := make([]byte, 64)
	rand.Read(val)

	// 填充 100万 key
	for i := 0; i < 1_000_000; i++ {
		v := make([]byte, 64)
		copy(v, val)
		m.Store(strconv.Itoa(i), v)
	}

	// 用 pprof 采集 3 秒 CPU profile，模拟持续读写负载
	var buf bytes.Buffer
	pprof.StartCPUProfile(&buf)

	// 3 秒内持续读写
	done := make(chan struct{})
	go func() {
		time.Sleep(3 * time.Second)
		close(done)
	}()

	workers := 8
	for w := 0; w < workers; w++ {
		go func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				select {
				case <-done:
					return
				default:
					key := strconv.Itoa(r.Intn(1_000_000))
					if r.Intn(100) < 70 {
						m.Load(key)
					} else {
						v := make([]byte, 64)
						m.Store(key, v)
					}
				}
			}
		}()
	}
	<-done
	time.Sleep(100 * time.Millisecond)
	pprof.StopCPUProfile()

	// 写入 pprof 文件
	os.WriteFile("/tmp/syncmap_cpu.prof", buf.Bytes(), 0644)
	fmt.Println("CPU profile 已写入 /tmp/syncmap_cpu.prof")
	fmt.Println("用 go tool pprof -top /tmp/syncmap_cpu.prof | grep gc 查看 GC 占比")

	// 同时直接输出 GC 统计
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("\nsync.Map (100万 key, 3秒负载):\n")
	fmt.Printf("  NumGC: %d\n", ms.NumGC)
	fmt.Printf("  GCCPUFraction: %.4f%%\n", ms.GCCPUFraction*100)
	fmt.Printf("  PauseTotalNs: %d µs\n", ms.PauseTotalNs/1000)
}
