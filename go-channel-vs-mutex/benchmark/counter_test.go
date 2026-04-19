package benchmark

import (
	"sync"
	"sync/atomic"
	"testing"
)

// ============================================================
// 场景1：计数器 — Channel vs Mutex vs Atomic
// 公平条件：100 goroutine 并发递增同一计数器，各执行 b.N 次
// ============================================================

// Mutex 保护计数器
func BenchmarkCounter_Mutex(b *testing.B) {
	var mu sync.Mutex
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			count++
			mu.Unlock()
		}
	})
	_ = count
}

// Channel 保护计数器（通过单一 writer goroutine 串行化写入）
func BenchmarkCounter_Channel(b *testing.B) {
	ch := make(chan struct{}, 1)
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- struct{}{}
			count++
			<-ch
		}
	})
	_ = count
}

// Atomic 计数器（作为基线参照）
func BenchmarkCounter_Atomic(b *testing.B) {
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&count, 1)
		}
	})
}

// ============================================================
// 不同竞争强度：固定 Mutex vs Channel，变化 GOMAXPROCS
// ============================================================

func benchmarkCounterMutexWithProcs(b *testing.B, procs int) {
	b.SetParallelism(procs)
	var mu sync.Mutex
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			count++
			mu.Unlock()
		}
	})
	_ = count
}

func benchmarkCounterChannelWithProcs(b *testing.B, procs int) {
	b.SetParallelism(procs)
	ch := make(chan struct{}, 1)
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- struct{}{}
			count++
			<-ch
		}
	})
	_ = count
}

func BenchmarkContention_Mutex_1(b *testing.B)    { benchmarkCounterMutexWithProcs(b, 1) }
func BenchmarkContention_Mutex_10(b *testing.B)   { benchmarkCounterMutexWithProcs(b, 10) }
func BenchmarkContention_Mutex_100(b *testing.B)  { benchmarkCounterMutexWithProcs(b, 100) }
func BenchmarkContention_Mutex_1000(b *testing.B) { benchmarkCounterMutexWithProcs(b, 1000) }

func BenchmarkContention_Channel_1(b *testing.B)    { benchmarkCounterChannelWithProcs(b, 1) }
func BenchmarkContention_Channel_10(b *testing.B)   { benchmarkCounterChannelWithProcs(b, 10) }
func BenchmarkContention_Channel_100(b *testing.B)  { benchmarkCounterChannelWithProcs(b, 100) }
func BenchmarkContention_Channel_1000(b *testing.B) { benchmarkCounterChannelWithProcs(b, 1000) }
