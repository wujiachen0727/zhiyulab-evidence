// E2a: 锁争用证伪实验（纯锁争用场景，移除 channel 干扰）
//
// 实验目的：
//   让 mutex profile 真正有数据，对比：
//   - CPU profile：能否看见锁等待？
//   - Mutex profile：能否看见锁争用？
//   - Trace：能否看见锁的因果关系？
//
// 关键设计：
//   - 延长临界区（让锁持有时间明显）
//   - 移除 channel 阻塞（消除"天然节流阀"）
//   - 增加 worker 数量（加大争用压力）
//
// 实测环境：Go 1.26.2 / darwin arm64
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"sync/atomic"
	"time"
)

const (
	numWorkers = 50            // 50 个 worker 争抢 1 把锁（加大压力）
	numTasks   = 200000        // 总任务数（加大任务量）
	critSize   = 5000          // 临界区工作量（大幅延长——约 5-10μs 级）
	cpuOutside = 500           // 锁外 CPU 工作量（保持较短，让锁争用成为主要耗时）
	experiment = 5 * time.Second
)

type shared struct {
	mu      sync.Mutex
	counter int64
	// 大一点的数组，增加临界区时间
	data [critSize]int
}

// incrementHeavy：故意延长临界区时间（多轮计算，约 5-10μs 级）
func (s *shared) incrementHeavy() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 临界区做更多工作，拉长锁持有时间
	// 3 轮 * 5000 次 = 15000 次写 + 混合运算
	for round := 0; round < 3; round++ {
		for i := 0; i < len(s.data); i++ {
			s.data[i] = (int(s.counter) ^ i) * 7
			s.data[i] = s.data[i]*13 + round
		}
	}
	s.counter++
}

// cpuOutsideLock：锁外 CPU 工作——让 CPU profile 里有业务代码出现
func cpuOutsideLock() int {
	sum := 0
	for i := 0; i < cpuOutside; i++ {
		sum += i*i - i
	}
	return sum
}

func runWorkload() {
	var wg sync.WaitGroup
	s := &shared{}
	var completed int64
	start := time.Now()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				if atomic.LoadInt64(&completed) >= int64(numTasks) {
					return
				}

				// 锁外 CPU 工作
				_ = cpuOutsideLock()

				// 锁内工作（重度争用）
				s.incrementHeavy()

				atomic.AddInt64(&completed, 1)
			}
		}(w)
	}

	timeout := time.After(experiment)
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-timeout:
		log.Printf("⚠️ 工作负载超时停止（这是预期的——锁争用会让完成速度变慢）")
	}

	elapsed := time.Since(start)
	done := atomic.LoadInt64(&completed)
	log.Printf("✅ 完成任务数: %d / %d, 耗时: %v", done, numTasks, elapsed)
	log.Printf("   平均每任务耗时: %v", elapsed/time.Duration(done))
}

func main() {
	log.Printf("=== E2a 锁争用证伪实验 ===")
	log.Printf("Go 版本: %s, GOMAXPROCS: %d, NumCPU: %d",
		runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())
	log.Printf("配置: %d workers, %d tasks, critSize=%d, cpuOutside=%d",
		numWorkers, numTasks, critSize, cpuOutside)

	// 开启所有 profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	cpuFile, err := os.Create("output/cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpuFile.Close()
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	traceFile, err := os.Create("output/trace.out")
	if err != nil {
		log.Fatal(err)
	}
	defer traceFile.Close()
	if err := trace.Start(traceFile); err != nil {
		log.Fatal(err)
	}
	defer trace.Stop()

	runWorkload()

	// 写 block / mutex profile
	blockFile, _ := os.Create("output/block.pprof")
	defer blockFile.Close()
	_ = pprof.Lookup("block").WriteTo(blockFile, 0)

	mutexFile, _ := os.Create("output/mutex.pprof")
	defer mutexFile.Close()
	_ = pprof.Lookup("mutex").WriteTo(mutexFile, 0)

	fmt.Println("✅ 全部 profile 已写入 output/")
}
