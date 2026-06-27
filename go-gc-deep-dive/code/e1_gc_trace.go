// E1: GC Trace 实测 — 典型工作负载下的 GC 行为
// 运行方式: GODEBUG=gctrace=1 go run main.go
// 预期输出: GC 频率、STW 时长、标记耗时等
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	// 打印初始内存状态
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("初始: Alloc=%d KB, Sys=%d KB, NumGC=%d\n", m.Alloc/1024, m.Sys/1024, m.NumGC)

	// 模拟典型 Web 服务工作负载：短生命周期对象 + 少量长期对象
	longLived := make([]*[]byte, 100) // 100 个长期存活对象
	for i := range longLived {
		b := make([]byte, 1024) // 每个 1KB
		longLived[i] = &b
	}

	// 持续分配短生命周期对象，模拟 HTTP 请求处理
	for i := 0; i < 1000000; i++ {
		_ = make([]byte, 256) // 短生命周期小对象

		// 每 1000 次分配做一次业务逻辑
		if i%1000 == 0 {
			runtime.GC() // 触发 GC 观察
		}
	}

	runtime.ReadMemStats(&m)
	fmt.Printf("结束: Alloc=%d KB, Sys=%d KB, NumGC=%d, PauseTotalNs=%d\n",
		m.Alloc/1024, m.Sys/1024, m.NumGC, m.PauseTotalNs)

	// 保持长期对象活跃
	_ = longLived
	time.Sleep(100 * time.Millisecond)
}
