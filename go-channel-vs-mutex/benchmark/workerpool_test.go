package benchmark

import (
	"runtime"
	"sync"
	"testing"
)

// ============================================================
// 场景3：工作池 — Channel 协调 worker
// 这是 Channel 的正确舞台：协调多个 worker 处理任务
// ============================================================

func BenchmarkWorkerPool_Channel(b *testing.B) {
	numWorkers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, numWorkers*2)
	var wg sync.WaitGroup

	// 启动 workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				// 模拟轻量计算任务
				_ = j * j
			}
		}()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

// 用 Mutex + 条件变量实现的"工作池"（反面对照，展示笨拙）
func BenchmarkWorkerPool_Mutex(b *testing.B) {
	numWorkers := runtime.GOMAXPROCS(0)
	var mu sync.Mutex
	var cond = sync.NewCond(&mu)
	queue := make([]int, 0, 1024)
	done := false
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				for len(queue) == 0 && !done {
					cond.Wait()
				}
				if done && len(queue) == 0 {
					mu.Unlock()
					return
				}
				j := queue[0]
				queue = queue[1:]
				mu.Unlock()
				_ = j * j
			}
		}()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		queue = append(queue, i)
		mu.Unlock()
		cond.Signal()
	}

	mu.Lock()
	done = true
	mu.Unlock()
	cond.Broadcast()
	wg.Wait()
}
