// E4：读多写少场景 —— sync.Mutex 读锁 vs sync.RWMutex 读锁（并发读）。
//
// 测量目标：在 GOMAXPROCS 个 goroutine 并发只读下，RWMutex 的 RLock 不互斥、
// 而 Mutex 的 Lock 串行化，对比两者吞吐。
// 关键前提：RWMutex 的优势只在"并发读"时出现；单 goroutine 下 RWMutex 反而更重。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem -cpu=14。
// 防 DCE：读到的值累加进 package-level sink。
package rwmutex

import (
	"sync"
	"testing"
)

var sinkInt64 int64

func BenchmarkMutexReadParallel(b *testing.B) {
	var mu sync.Mutex
	var data int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			sinkInt64 += data
			mu.Unlock()
		}
	})
}

func BenchmarkRWMutexReadParallel(b *testing.B) {
	var mu sync.RWMutex
	var data int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			sinkInt64 += data
			mu.RUnlock()
		}
	})
}

// 读临界区内做真实工作量（对 1KB 切片求和）——此时 Mutex 串行化的代价超过
// RWMutex 的 readerCount 缓存行争用，RWMutex 才显现"读不互斥"的优势。
var readBuf = make([]int64, 128)

func BenchmarkMutexReadWorkParallel(b *testing.B) {
	var mu sync.Mutex
	var sum int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			s := int64(0)
			for _, v := range readBuf {
				s += v
			}
			sum = s
			mu.Unlock()
		}
	})
	sinkInt64 += sum
}

func BenchmarkRWMutexReadWorkParallel(b *testing.B) {
	var mu sync.RWMutex
	var sum int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			s := int64(0)
			for _, v := range readBuf {
				s += v
			}
			sum = s
			mu.RUnlock()
		}
	})
	sinkInt64 += sum
}
