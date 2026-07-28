package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)

	// 创建未触发的 Timer（1小时超时），不保存任何引用
	// 这些 Timer 注册在 runtime timer heap 中
	// Go 1.22: Timer 在 runtime heap 中不会被 GC 回收
	// Go 1.23+: 未引用的未触发 Timer 可以被 GC 回收
	for i := 0; i < 100000; i++ {
		// time.After 返回 channel，我们不保存引用
		// 但 channel 本身可能被 Go runtime 内部引用
		_ = time.After(1 * time.Hour)
	}

	// 多次 GC
	for j := 0; j < 10; j++ {
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
	}

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	allocInuse := int64(m1.HeapInuse) - int64(m0.HeapInuse)
	objInuse := int64(m1.HeapObjects) - int64(m0.HeapObjects)

	fmt.Printf("=== time.After 泄漏实验（未触发 Timer，纯引用计数）===\n")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("循环次数: 100000\n")
	fmt.Printf("HeapInuse 增量: %d bytes (%.2f MB)\n", allocInuse, float64(allocInuse)/1024/1024)
	fmt.Printf("HeapObjects 增量: %d\n", objInuse)
	fmt.Printf("NumGC: %d\n", m1.NumGC-m0.NumGC)
	
	// 额外信息：检查 runtime 中的 timer 数量
	// 通过 NumGoroutine 间接观察（timer 不增加 goroutine，但可以看 GC 效果）
	fmt.Printf("\n注：如果 Go 1.23 的 HeapObjects 显著低于 Go 1.22，说明 GC 回收了未触发的 Timer")
}
