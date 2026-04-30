package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/ristretto"
)

// E1: sync.Map vs bigcache vs ristretto 在不同读写比下的 benchmark
// 测试场景：10万个 key，value 64 bytes，不同读写比
// 目的：证明 sync.Map 在写密集场景下性能骤降

const (
	keyCount  = 100_000
	valueSize = 64
)

var (
	sinkBytes []byte
	sinkBool  bool
)

func generateValue() []byte {
	b := make([]byte, valueSize)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	return b
}

// ==================== sync.Map ====================

func BenchmarkSyncMap_Read90_Write10(b *testing.B) {
	benchSyncMap(b, 90)
}

func BenchmarkSyncMap_Read70_Write30(b *testing.B) {
	benchSyncMap(b, 70)
}

func BenchmarkSyncMap_Read50_Write50(b *testing.B) {
	benchSyncMap(b, 50)
}

func benchSyncMap(b *testing.B, readPercent int) {
	var m sync.Map
	// 预填充
	for i := 0; i < keyCount; i++ {
		m.Store(strconv.Itoa(i), generateValue())
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(r.Intn(keyCount))
			if r.Intn(100) < readPercent {
				v, _ := m.Load(key)
				if v != nil {
					sinkBytes = v.([]byte)
				}
			} else {
				m.Store(key, generateValue())
			}
		}
	})
}

// ==================== bigcache ====================

func BenchmarkBigcache_Read90_Write10(b *testing.B) {
	benchBigcache(b, 90)
}

func BenchmarkBigcache_Read70_Write30(b *testing.B) {
	benchBigcache(b, 70)
}

func BenchmarkBigcache_Read50_Write50(b *testing.B) {
	benchBigcache(b, 50)
}

func benchBigcache(b *testing.B, readPercent int) {
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	// 预填充
	for i := 0; i < keyCount; i++ {
		cache.Set(strconv.Itoa(i), generateValue())
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(r.Intn(keyCount))
			if r.Intn(100) < readPercent {
				v, _ := cache.Get(key)
				sinkBytes = v
			} else {
				cache.Set(key, generateValue())
			}
		}
	})
}

// ==================== ristretto ====================

func BenchmarkRistretto_Read90_Write10(b *testing.B) {
	benchRistretto(b, 90)
}

func BenchmarkRistretto_Read70_Write30(b *testing.B) {
	benchRistretto(b, 70)
}

func BenchmarkRistretto_Read50_Write50(b *testing.B) {
	benchRistretto(b, 50)
}

func benchRistretto(b *testing.B, readPercent int) {
	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,
		MaxCost:     1 << 28, // 256MB
		BufferItems: 64,
	})
	// 预填充
	for i := 0; i < keyCount; i++ {
		cache.Set(strconv.Itoa(i), generateValue(), int64(valueSize))
	}
	cache.Wait()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(r.Intn(keyCount))
			if r.Intn(100) < readPercent {
				v, found := cache.Get(key)
				if found && v != nil {
					sinkBytes = v.([]byte)
				}
				sinkBool = found
			} else {
				cache.Set(key, generateValue(), int64(valueSize))
			}
		}
	})
}

// ==================== 主入口（说明用途）====================

func main() {
	fmt.Println("请使用 go test -bench=. -benchmem -cpu=8 运行 benchmark")
	fmt.Println("测试场景：10万 key，64B value，不同读写比")
}
