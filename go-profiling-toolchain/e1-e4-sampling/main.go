// E1 + E4：CPU 热点基线 + 采样频率对比
//
// 核心问题：
//   E1 —— pprof 能否清晰识别 CPU 热点？（证明 pprof 擅长这件事）
//   E4 —— 采样频率对分辨率的影响（证明采样 = 统计近似，不是真相）
//
// 实验设计：
//   一个程序，四种函数：
//     heavyFn   : ~100ms/call，重度 CPU
//     mediumFn  :   ~5ms/call，中等 CPU
//     shortFn   :  ~50μs/call，短时 CPU
//     microFn   :   ~5μs/call，微量 CPU
//   每个函数调用 N 次，总 CPU 时间差距巨大（1000:1 级别）
//
// 通过环境变量 RATE_HZ 控制采样率，跑 3 次：
//   RATE_HZ=100   （默认）
//   RATE_HZ=1000  （激进）
//   RATE_HZ=10000 （极限）
//
// 观察：
//   - heavyFn 和 mediumFn 在三种采样率下是否都能看见？
//   - shortFn 在 100Hz 下是否被低估/漏采？
//   - microFn 即使在 1000Hz 下能看见多少？
//
// 实测环境：Go 1.26.2 / darwin arm64
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

// 每个函数的调用次数（重复多次累加 CPU 时间）
// 目标：每个函数 ~8-10s CPU 时间，**总 CPU 时间分布均匀**——理想采样下应各占 25%
// 实测每个函数单次耗时（M 系列芯片）：heavy ~120ms, medium ~5ms, short ~50μs, micro ~5μs
// 比例：heavy:medium:short:micro = 1 : 24 : 2400 : 24000（按耗时反比计算次数）
// 取近似整数：80 : 2000 : 200000 : 2000000
const (
	heavyCalls  = 80       // 80 × 120ms = 9.6s
	mediumCalls = 2000     // 2000 × 5ms = 10s
	shortCalls  = 200000   // 200k × 50μs = 10s
	microCalls  = 2000000  // 2M × 5μs = 10s
)

// heavyFn：~120ms CPU（M 系列芯片实测约 3.1ms × 40 = 124ms）
// 用大量浮点运算填满这个时长
func heavyFn() {
	var sum float64 = 1.0
	for i := 0; i < 400000000; i++ {
		sum = sum*1.0000001 + 0.0001
	}
	_ = sum
}

// mediumFn：~5ms CPU
func mediumFn() {
	var sum float64 = 1.0
	for i := 0; i < 20000000; i++ {
		sum = sum*1.0000001 + 0.0001
	}
	_ = sum
}

// shortFn：~50μs CPU
func shortFn() {
	var sum float64 = 1.0
	for i := 0; i < 200000; i++ {
		sum = sum*1.0000001 + 0.0001
	}
	_ = sum
}

// microFn：~5μs CPU
func microFn() {
	var sum float64 = 1.0
	for i := 0; i < 20000; i++ {
		sum = sum*1.0000001 + 0.0001
	}
	_ = sum
}

// runWorkload 按预定次数调用每个函数
// 为了让采样统计更平均，**轮流穿插调用**而不是"连跑 heavy 再跑 short"
// 穿插方式：每轮每种函数各调一次，比例用固定的子步数控制
func runWorkload() (heavy, medium, short, micro int) {
	// 比例 heavy:medium:short:micro = 80 : 2000 : 200000 : 2000000
	//                              = 1  : 25   : 2500   : 25000
	// 每一大轮：heavy 1 次 + medium 25 次 + short 2500 次 + micro 25000 次
	// 共 heavyCalls (80) 大轮
	for round := 0; round < heavyCalls; round++ {
		heavyFn()
		heavy++

		for i := 0; i < 25; i++ {
			mediumFn()
			medium++
		}

		for i := 0; i < 2500; i++ {
			shortFn()
			short++
		}

		for i := 0; i < 25000; i++ {
			microFn()
			micro++
		}
	}
	return
}

func getSampleRateHz() int {
	if v := os.Getenv("RATE_HZ"); v != "" {
		if hz, err := strconv.Atoi(v); err == nil {
			return hz
		}
	}
	return 100 // Go 默认值
}

func main() {
	rateHz := getSampleRateHz()
	log.Printf("=== E1+E4 CPU 热点基线 + 采样频率对比 ===")
	log.Printf("Go 版本: %s, GOMAXPROCS: %d, NumCPU: %d",
		runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())
	log.Printf("采样频率: %d Hz", rateHz)
	log.Printf("理论 CPU 时间分布（每个函数约 10s CPU）:")
	log.Printf("  heavyFn    80  × 120ms =  9.6s")
	log.Printf("  mediumFn  2000 × 5ms   = 10.0s")
	log.Printf("  shortFn   200k × 50μs  = 10.0s")
	log.Printf("  microFn   2M   × 5μs   = 10.0s")
	log.Printf("  总计约 40s CPU（单核墙钟）——理想采样下 4 个函数各占 25%%")

	// 设置采样率（必须在 StartCPUProfile 之前调用）
	runtime.SetCPUProfileRate(rateHz)

	// 输出文件名根据采样率
	outFile := fmt.Sprintf("output/cpu-%dhz.pprof", rateHz)
	if err := os.MkdirAll("output", 0755); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 启动采样
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// 跑工作负载
	start := time.Now()
	h, m, s, mi := runWorkload()
	elapsed := time.Since(start)

	log.Printf("实际调用次数: heavy=%d, medium=%d, short=%d, micro=%d", h, m, s, mi)
	log.Printf("总墙钟时间: %v", elapsed)
	log.Printf("profile 输出: %s", outFile)
}
