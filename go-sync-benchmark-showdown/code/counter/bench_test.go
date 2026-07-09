// E2：高频计数场景 —— atomic.AddInt64 vs sync.Mutex 自增。
//
// 测量目标：单 goroutine 下，1000 万次自增的每次操作开销（无并发争用，
// 对比的是"抽象高级的锁"与"无锁原子"的纯路径差异）。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem。
// 防 DCE：累加结果赋给 package-level sink。
package counter

import (
	"sync"
	"sync/atomic"
	"testing"
)

var sinkInt64 int64

func BenchmarkCounterAtomic(b *testing.B) {
	var c int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&c, 1)
	}
	sinkInt64 += c
}

func BenchmarkCounterMutex(b *testing.B) {
	var mu sync.Mutex
	var c int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		c++
		mu.Unlock()
	}
	sinkInt64 += c
}

// BenchmarkCounterAtomicParallel / MutexParallel：GOMAXPROCS 个 goroutine 并发
// 争用同一个计数器——这才是"高频计数"的真实场景，atomic 的无锁优势在此显现。
func BenchmarkCounterAtomicParallel(b *testing.B) {
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&c, 1)
		}
	})
}

func BenchmarkCounterMutexParallel(b *testing.B) {
	var mu sync.Mutex
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			c++
			mu.Unlock()
		}
	})
}
