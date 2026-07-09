// E3：批量并发编排场景 —— sync.WaitGroup 在 worker 数 100/1000/10000 下的每批开销。
//
// 测量目标：每批启动 N 个 goroutine 做"空工作"并 Wait 回收，观察 ns/op 随 worker 数增长。
// 注意：ns/op 是"整批"的开销，不是单 goroutine；重点看退化曲线（线性 vs 超线性）。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem。
// 防 DCE：goroutine 内对 sink 做自增，避免被优化掉。
package waitgroup

import (
	"sync"
	"testing"
)

var sinkInt int

func benchmarkWaitGroupN(b *testing.B, workers int) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(workers)
		for w := 0; w < workers; w++ {
			go func() {
				defer wg.Done()
				sinkInt++
			}()
		}
		wg.Wait()
	}
}

func BenchmarkWaitGroup100(b *testing.B)   { benchmarkWaitGroupN(b, 100) }
func BenchmarkWaitGroup1000(b *testing.B)  { benchmarkWaitGroupN(b, 1000) }
func BenchmarkWaitGroup10000(b *testing.B) { benchmarkWaitGroupN(b, 10000) }
