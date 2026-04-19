// E2 等待陷阱证伪实验
//
// 假设（待证伪）：pprof CPU profile 无法识别"等待导致的慢"
//
// 实验设计：
//   - 构造一个故意让 goroutine 大量时间花在锁等待上的场景
//   - 少量 CPU 计算（让 CPU profile 有东西可看但不是瓶颈）
//   - 大量锁争用（真正的瓶颈）
//   - 同时开启 CPU profile / block profile / mutex profile / execution trace
//   - 对比四种 profile 看到的"故事"
//
// 运行方式：
//   go run main.go
//   ls -la output/  # 产出 cpu.pprof / block.pprof / mutex.pprof / trace.out
//
// 分析命令：
//   go tool pprof -top output/cpu.pprof
//   go tool pprof -top output/block.pprof
//   go tool pprof -top output/mutex.pprof
//   go tool trace output/trace.out
//
// 实测环境：Go 1.26.2 / darwin arm64 / M 系列芯片
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
	numWorkers = 20       // 20 个 worker 争抢同一把锁（制造高竞争）
	numTasks   = 4000     // 总任务数
	cpuWorkMu  = 50       // 每个任务的 CPU 计算强度（微乎其微）
	experiment = 3 * time.Second
)

// shared 是所有 worker 共享的受锁保护的状态
// 故意设计成"一次只能一个 worker 访问"——制造重度锁争用
type shared struct {
	mu      sync.Mutex
	counter int64
	data    [1024]int // 一点点内存，让锁临界区有东西做
}

// incrementWithLock 在锁内做非常少量的工作
// 这是本实验的"等待陷阱"核心：每个 goroutine 大部分时间在等锁
func (s *shared) incrementWithLock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 锁内做一点点工作（让锁持有时间有一定长度，强化争用）
	for i := 0; i < len(s.data); i++ {
		s.data[i] = int(s.counter) + i
	}
	s.counter++
}

// cpuWork 是故意制造的 CPU 计算任务
// 故意做得很短——让 CPU profile 看起来"有事在干"但不是主要耗时点
func cpuWork() int {
	sum := 0
	for i := 0; i < cpuWorkMu; i++ {
		// 防止编译器优化掉
		sum += i * i
	}
	return sum
}

// blockOnChannel 模拟下游慢的场景（channel 阻塞）
func blockOnChannel(ch chan int) {
	// 等下游返回一个结果——下游故意慢
	<-ch
}

// slowDownstream 作为下游服务，对每个请求延迟一点点
func slowDownstream(ch chan int, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			// 每次被唤醒就睡一小会儿，再返回结果
			time.Sleep(200 * time.Microsecond)
			select {
			case ch <- 1:
			case <-done:
				return
			}
		}
	}
}

func runWorkload() {
	var wg sync.WaitGroup
	s := &shared{}
	ch := make(chan int, numWorkers) // 带缓冲避免完全卡死
	done := make(chan struct{})

	// 启动慢下游
	go slowDownstream(ch, done)

	// 原子计数：完成的任务数
	var completed int64
	start := time.Now()

	// 启动 numWorkers 个 worker 争抢同一把锁
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				if atomic.LoadInt64(&completed) >= int64(numTasks) {
					return
				}

				// CPU 工作（微量）
				_ = cpuWork()

				// 锁等待（重度争用，本实验的主要等待来源）
				s.incrementWithLock()

				// channel 阻塞（次要等待来源，证明"等待"不止一种）
				blockOnChannel(ch)

				atomic.AddInt64(&completed, 1)
			}
		}(w)
	}

	// 总超时保护
	timeout := time.After(experiment)
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-timeout:
		log.Printf("⚠️ 工作负载超过 %v 超时，停止采集", experiment)
	}

	close(done) // 通知下游退出
	elapsed := time.Since(start)
	log.Printf("✅ 完成任务数: %d, 耗时: %v", atomic.LoadInt64(&completed), elapsed)
}

func main() {
	log.Printf("=== E2 等待陷阱证伪实验 ===")
	log.Printf("Go 版本: %s, GOMAXPROCS: %d, NumCPU: %d",
		runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())

	// 开启所有相关的 profiling
	// 1. block profile: 记录 goroutine 被阻塞的情况（锁、channel、select 等）
	//    参数是"采样率"——1 表示每个阻塞事件都记录，适合小型实验
	runtime.SetBlockProfileRate(1)

	// 2. mutex profile: 记录 mutex 持有/争用情况
	//    参数是"采样频率分母"——1 表示每次都采样
	runtime.SetMutexProfileFraction(1)

	// 3. CPU profile: 默认 100Hz 采样
	cpuFile, err := os.Create("output/cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpuFile.Close()
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// 4. execution trace: 完整事件追踪
	traceFile, err := os.Create("output/trace.out")
	if err != nil {
		log.Fatal(err)
	}
	defer traceFile.Close()
	if err := trace.Start(traceFile); err != nil {
		log.Fatal(err)
	}
	defer trace.Stop()

	// 运行工作负载
	runWorkload()

	// 写入 block profile
	blockFile, err := os.Create("output/block.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer blockFile.Close()
	if err := pprof.Lookup("block").WriteTo(blockFile, 0); err != nil {
		log.Fatal(err)
	}

	// 写入 mutex profile
	mutexFile, err := os.Create("output/mutex.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer mutexFile.Close()
	if err := pprof.Lookup("mutex").WriteTo(mutexFile, 0); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ 全部 profile 已写入 output/ 目录")
	fmt.Println("")
	fmt.Println("分析命令：")
	fmt.Println("  go tool pprof -top output/cpu.pprof")
	fmt.Println("  go tool pprof -top output/block.pprof")
	fmt.Println("  go tool pprof -top output/mutex.pprof")
	fmt.Println("  go tool trace output/trace.out")
}
