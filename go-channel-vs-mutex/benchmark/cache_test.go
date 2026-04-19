package benchmark

import (
	"sync"
	"testing"
)

// ============================================================
// 场景2：缓存（map 保护）— Channel vs RWMutex
// 模拟 90% 读 + 10% 写的典型缓存场景
// ============================================================

// RWMutex 保护 map
func BenchmarkCache_RWMutex(b *testing.B) {
	var mu sync.RWMutex
	cache := make(map[string]string)
	cache["key"] = "value"

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if i%10 == 0 {
				// 10% 写
				mu.Lock()
				cache["key"] = "new_value"
				mu.Unlock()
			} else {
				// 90% 读
				mu.RLock()
				_ = cache["key"]
				mu.RUnlock()
			}
		}
	})
}

// Channel 保护 map（所有操作串行化到单一 goroutine）
func BenchmarkCache_Channel(b *testing.B) {
	type cacheOp struct {
		write bool
		key   string
		value string
		resp  chan string
	}

	ch := make(chan cacheOp, 64)
	cache := make(map[string]string)
	cache["key"] = "value"

	// 单一 cache manager goroutine
	go func() {
		for op := range ch {
			if op.write {
				cache[op.key] = op.value
				op.resp <- ""
			} else {
				op.resp <- cache[op.key]
			}
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		resp := make(chan string, 1)
		i := 0
		for pb.Next() {
			i++
			if i%10 == 0 {
				ch <- cacheOp{write: true, key: "key", value: "new_value", resp: resp}
				<-resp
			} else {
				ch <- cacheOp{write: false, key: "key", resp: resp}
				<-resp
			}
		}
	})

	close(ch)
}
