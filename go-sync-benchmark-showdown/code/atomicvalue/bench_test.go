// E5：读多写少场景 —— atomic.Value 无锁读 vs sync.RWMutex 读锁（并发读）。
//
// 测量目标：GOMAXPROCS 个 goroutine 并发只读下，atomic.Value.Load（无锁）与
// RWMutex.RLock（仍有原子计数）的每次读开销对比。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem -cpu=14。
// 防 DCE：读到的值累加进 package-level sink。
package atomicvalue

import (
	"sync"
	"sync/atomic"
	"testing"
)

var sinkInt64 int64

func BenchmarkAtomicValueReadParallel(b *testing.B) {
	var v atomic.Value
	v.Store(int64(42))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sinkInt64 += v.Load().(int64)
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
