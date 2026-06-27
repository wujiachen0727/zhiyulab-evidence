// Benchmark — 协程池 vs 直接 go func() 性能对比
// 运行环境：Go 1.26.4, darwin/arm64
// 覆盖多量级（10/100/1000/10000 任务）和两种任务类型（CPU 密集 / I/O 模拟）

package benchmark

import (
	"sync"
	"testing"
)

// ---------- 任务定义 ----------

// cpuIntensiveTask 执行 CPU 密集型计算
func cpuIntensiveTask() {
	var sum int
	for i := 0; i < 10000; i++ {
		sum += i * i
	}
	_ = sum
}

// ioSimulatedTask 模拟 I/O 等待（如数据库查询、HTTP 调用）
func ioSimulatedTask() {
	// 使用纯计算模拟 I/O 等待（避免 time.Sleep 的 goroutine 调度偏差）
	var sum int
	for i := 0; i < 500000; i++ {
		sum += i
	}
	_ = sum
}

// ---------- 协程池实现 ----------

// Pool 是一个简单的 goroutine worker pool
type Pool struct {
	taskCh chan func()
	wg     sync.WaitGroup
}

// NewPool 创建一个指定 worker 数量的协程池
func NewPool(workerCount, taskQueueSize int) *Pool {
	p := &Pool{
		taskCh: make(chan func(), taskQueueSize),
	}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for task := range p.taskCh {
				task()
			}
		}()
	}
	return p
}

// Submit 提交一个任务到池中
func (p *Pool) Submit(task func()) {
	p.taskCh <- task
}

// Shutdown 等待所有任务完成并关闭池
func (p *Pool) Shutdown() {
	close(p.taskCh)
	p.wg.Wait()
}

// ---------- Benchmark 辅助函数 ----------

func runDirect(n int, task func()) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task()
		}()
	}
	wg.Wait()
}

func runWithPool(n int, workers int, task func()) {
	pool := NewPool(workers, n)
	for i := 0; i < n; i++ {
		pool.Submit(task)
	}
	pool.Shutdown()
}

// ---------- CPU 密集任务 Benchmark ----------

func BenchmarkCPUDirect10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(10, cpuIntensiveTask)
	}
}

func BenchmarkCPUPool10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(10, 8, cpuIntensiveTask)
	}
}

func BenchmarkCPUDirect100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(100, cpuIntensiveTask)
	}
}

func BenchmarkCPUPool100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(100, 8, cpuIntensiveTask)
	}
}

func BenchmarkCPUDirect1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(1000, cpuIntensiveTask)
	}
}

func BenchmarkCPUPool1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 8, cpuIntensiveTask)
	}
}

func BenchmarkCPUDirect10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(10000, cpuIntensiveTask)
	}
}

func BenchmarkCPUPool10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(10000, 8, cpuIntensiveTask)
	}
}

// ---------- I/O 模拟任务 Benchmark ----------

func BenchmarkIODirect10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(10, ioSimulatedTask)
	}
}

func BenchmarkIOPool10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(10, 8, ioSimulatedTask)
	}
}

func BenchmarkIODirect100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(100, ioSimulatedTask)
	}
}

func BenchmarkIOPool100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(100, 8, ioSimulatedTask)
	}
}

func BenchmarkIODirect1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(1000, ioSimulatedTask)
	}
}

func BenchmarkIOPool1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 8, ioSimulatedTask)
	}
}

func BenchmarkIODirect10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDirect(10000, ioSimulatedTask)
	}
}

func BenchmarkIOPool10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(10000, 8, ioSimulatedTask)
	}
}

// ---------- Worker 数量影响 Benchmark ----------

func BenchmarkPoolWorkers1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 1, cpuIntensiveTask)
	}
}

func BenchmarkPoolWorkers4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 4, cpuIntensiveTask)
	}
}

func BenchmarkPoolWorkers8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 8, cpuIntensiveTask)
	}
}

func BenchmarkPoolWorkers16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 16, cpuIntensiveTask)
	}
}

func BenchmarkPoolWorkers32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithPool(1000, 32, cpuIntensiveTask)
	}
}
