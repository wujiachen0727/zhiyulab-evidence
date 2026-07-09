// E1：单例获取开销对比 —— sync.Once vs Channel（buffered chan，后台生产者持续补货）
//
// 测量目标：在"已初始化完成、反复获取单例"的场景下，两种机制的每次获取开销。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem。
// 防 DCE：所有返回值赋给 package-level sink。
package singleton

import (
	"sync"
	"testing"
)

var (
	sinkOnce int
	sinkChan int
)

// BenchmarkSingletonOnce 反复调用 once.Do 获取已初始化的单例。
// 第一次后 once.Do 内部的 func 不再执行，每次仅做一次 atomic 读 + 比较。
func BenchmarkSingletonOnce(b *testing.B) {
	var once sync.Once
	var v int
	once.Do(func() { v = 42 })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		once.Do(func() { v = 42 })
		sinkOnce += v
	}
}

// BenchmarkSingletonChannel 每次获取都从 buffered chan(1) 接收；
// 后台 goroutine 持续补货，使接收方总是能立即拿到，隔离出"接收"本身的开销。
func BenchmarkSingletonChannel(b *testing.B) {
	ch := make(chan int, 1)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case ch <- 42:
			case <-stop:
				return
			}
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkChan += <-ch
	}
	b.StopTimer()
	close(stop)
}

// BenchmarkSingletonChannelRoundtrip 同一 goroutine 内 buffered chan(1) 的
// 发送+接收往返，隔离出"通道操作本身"的纯开销（无独立生产者调度）。
func BenchmarkSingletonChannelRoundtrip(b *testing.B) {
	ch := make(chan int, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- 42
		sinkChan += <-ch
	}
}
