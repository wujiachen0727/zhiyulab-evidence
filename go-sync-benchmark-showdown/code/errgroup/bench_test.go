// E6：批量并发编排场景 —— 裸 sync.WaitGroup vs errgroup.Group（带 context 取消 + 错误聚合）。
//
// 测量目标：每批启动 10 个 goroutine 做"空工作"，对比 WaitGroup 与 errgroup 的每批开销。
// errgroup 额外付出的：context 传播、sync.Once 记录首个错误、互斥锁保护错误——即为"取消+聚合"能力付费。
// 测量条件：Go 1.26.4 darwin/arm64，GOMAXPROCS=14，go test -bench -benchmem。
// 防 DCE：goroutine 内对 sink 做自增。
package errgroup

import (
	"context"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
)

var sinkInt int

func BenchmarkWaitGroupPlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(10)
		for w := 0; w < 10; w++ {
			go func() {
				defer wg.Done()
				sinkInt++
			}()
		}
		wg.Wait()
	}
}

func BenchmarkErrgroup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g, _ := errgroup.WithContext(context.Background())
		for w := 0; w < 10; w++ {
			g.Go(func() error {
				sinkInt++
				return nil
			})
		}
		_ = g.Wait()
	}
}
