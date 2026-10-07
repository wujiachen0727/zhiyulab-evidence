package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// sink 防止编译器把忙等循环整体优化掉
var sink uint64

// busyParallel 启动 n 个 CPU 密集协程，返回全部完成所需的墙钟时间
func busyParallel(n int, iters int) time.Duration {
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var x uint64
			for j := 0; j < iters; j++ {
				x += uint64(j)
			}
			sink = x
		}()
	}
	wg.Wait()
	return time.Since(start)
}

// bestOf 重复跑 reps 次，取最小值。
// 计时类实验受调度抖动与机器负载影响，单次结果不可复现；
// 取最小值是最接近"这一档设置的真实下界"的稳定估计。
func bestOf(n int, iters int, reps int) time.Duration {
	best := time.Duration(1<<62 - 1)
	for i := 0; i < reps; i++ {
		if d := busyParallel(n, iters); d < best {
			best = d
		}
	}
	return best
}

func main() {
	fmt.Printf("runtime.Version()      = %s\n", runtime.Version())
	fmt.Printf("runtime.NumCPU()       = %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS(0)  = %d\n", runtime.GOMAXPROCS(0))

	fmt.Println()
	fmt.Println("== GOMAXPROCS 对并行度的影响 ==")
	fmt.Println("  4 个 CPU 密集协程，每个 1.5 亿次累加；每档先热身一次，再取 3 次的最小值")

	const (
		workers  = 4
		iters    = 150_000_000
		reps     = 3
		warmupIt = 15_000_000
	)

	original := runtime.GOMAXPROCS(0)
	results := make(map[int]time.Duration)
	for _, p := range []int{1, 2, 4, 8} {
		runtime.GOMAXPROCS(p)
		busyParallel(workers, warmupIt)
		d := bestOf(workers, iters, reps)
		results[p] = d
		fmt.Printf("  GOMAXPROCS=%d -> %v\n", p, d.Round(time.Millisecond))
	}
	runtime.GOMAXPROCS(original)

	base := results[1]
	fmt.Println()
	fmt.Println("  相对 GOMAXPROCS=1 的加速比：")
	for _, p := range []int{1, 2, 4, 8} {
		fmt.Printf("    GOMAXPROCS=%d -> %.2fx\n", p, float64(base)/float64(results[p]))
	}
}
