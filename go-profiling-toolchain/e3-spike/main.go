// E3 偶发毛刺证伪实验 —— HTTP 服务
//
// 假设（待证伪）：在 30 秒 pprof 采样窗口下，偶发毛刺会被平均稀释——
//   pprof 无法告诉你"异常发生在哪个时间点"。
//
// 服务设计：
//   - /fast：99% 的请求走这里。快速响应（1-5ms 的 CPU 工作）。
//   - /slow：1% 的请求走这里。慢路径（CPU 密集的特定函数）。
//   - 慢路径只在特定时间窗口集中发生（模拟"偶发毛刺"）——
//     不是每次请求 1% 概率，而是在第 15-20 秒这个特定窗口集中触发。
//
// 关键点：慢路径函数名特殊（processSlowPath），方便在 profile 里识别它的占比
//
// 运行方式：
//   go run main.go
//
// 对应的压测和采样脚本：run-experiment.sh
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof" // 暴露 /debug/pprof/ 接口
	"sync/atomic"
	"time"
)

// 服务启动时间（用于计算"当前第几秒"）
var startTime = time.Now()

// 已处理请求数
var reqCount int64
var slowCount int64

// fastHandler：快路径——约 1-5ms CPU 工作
func fastHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&reqCount, 1)
	processFastPath()
	fmt.Fprintln(w, "ok")
}

// slowHandler：慢路径——约 50-100ms CPU 工作
// 只在"毛刺窗口"（第 15-20 秒）集中触发
func slowHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&reqCount, 1)

	// 判断：当前时间是否在毛刺窗口内？
	elapsed := time.Since(startTime)
	inSpikeWindow := elapsed >= 15*time.Second && elapsed <= 20*time.Second

	if inSpikeWindow {
		// 进入毛刺：走慢路径
		atomic.AddInt64(&slowCount, 1)
		processSlowPath()
	} else {
		// 非毛刺窗口：/slow 也走快路径（模拟"99% 正常时这个路径也快"）
		processFastPath()
	}

	fmt.Fprintln(w, "ok")
}

// processFastPath：快路径 CPU 工作（约 15μs——正常 HTTP 请求的小量工作）
func processFastPath() {
	sum := 0
	for i := 0; i < 50000; i++ {
		sum += i * i
	}
	_ = sum
}

// processSlowPath：慢路径 CPU 工作（约 200ms——真实生产毛刺的典型量级）
// 独特的函数名让它在 profile 里容易识别
//
// 注意：慢路径耗时必须明显大于快路径，且足够长才能在 profile 里凸显
// 基准测试（M 系列芯片）：
//   - 200 万次乘法: ~500μs（不够明显）
//   - 8 亿次乘法: ~200ms（真实生产毛刺量级，profile 清晰可见）
func processSlowPath() {
	// 用 int64 避免溢出，且和 fastPath 的函数体结构明显不同
	// 便于 profile 符号表识别
	var sum int64
	for outer := 0; outer < 400; outer++ {
		for i := 0; i < 2000000; i++ {
			sum += int64(i) * int64(i) * 3
		}
	}
	_ = sum
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	elapsed := time.Since(startTime)
	fmt.Fprintf(w, "elapsed: %v\n", elapsed)
	fmt.Fprintf(w, "total requests: %d\n", atomic.LoadInt64(&reqCount))
	fmt.Fprintf(w, "slow path hits: %d\n", atomic.LoadInt64(&slowCount))
}

func main() {
	_ = rand.Int() // 消除 import 警告（某些 Go 版本）

	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/stats", statsHandler)

	log.Printf("=== E3 毛刺服务启动 ===")
	log.Printf("监听: :6060")
	log.Printf("端点:")
	log.Printf("  /fast        - 快路径（99%% 流量）")
	log.Printf("  /slow        - 慢路径（1%% 流量，仅在 15-20s 毛刺窗口内真正慢）")
	log.Printf("  /stats       - 服务状态")
	log.Printf("  /debug/pprof - pprof 接口")
	log.Printf("毛刺窗口: 第 15-20 秒")

	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Fatal(err)
	}
}
