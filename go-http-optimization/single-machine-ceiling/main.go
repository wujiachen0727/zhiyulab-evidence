// single-machine-ceiling: Go HTTP 服务业务复杂度衰减实验
// [实测 Go 1.26.2 darwin/arm64]
//
// 测试 Go HTTP 服务在不同业务复杂度下的 QPS/P99/GC 停顿衰减曲线，
// 揭示"Go 单机到底多强"在真实业务复杂度下的衰减。
//
// 5 层业务复杂度递进：
//   Layer 0: 纯 JSON 序列化
//   Layer 1: +模拟DB查询（2ms sleep）
//   Layer 2: +模拟Redis缓存（0.5ms sleep，先查缓存 miss 后查 DB）
//   Layer 3: +模拟外部HTTP调用（调用辅助服务器，3ms）
//   Layer 4: +业务逻辑（1ms 字符串处理+排序）

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 响应结构
// ============================================================

type Response struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

// ============================================================
// 辅助 HTTP 服务器（模拟外部调用）
// ============================================================

func startHelperServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/external", func(w http.ResponseWriter, r *http.Request) {
		// 模拟外部服务 3ms 处理时间
		time.Sleep(3 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"source":  "external-service",
			"version": "1.0",
			"region":  "us-east-1",
		})
	})
	go http.ListenAndServe(":"+port, mux)
	// 等待辅助服务器启动
	time.Sleep(200 * time.Millisecond)
}

// ============================================================
// 主服务器：5 层路由
// ============================================================

func layer0Handler(w http.ResponseWriter, r *http.Request) {
	// Layer 0: 纯 JSON 序列化
	resp := Response{
		Status: "ok",
		Data: map[string]string{
			"layer": "0",
			"desc":  "pure-json",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func layer1Handler(w http.ResponseWriter, r *http.Request) {
	// Layer 1: +模拟DB查询 2ms
	time.Sleep(2 * time.Millisecond) // [推演：模拟延迟]
	resp := Response{
		Status: "ok",
		Data: map[string]string{
			"layer": "1",
			"desc":  "db-query",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func layer2Handler(w http.ResponseWriter, r *http.Request) {
	// Layer 2: +模拟Redis缓存
	// 先查缓存（0.5ms），miss 后查 DB（2ms）
	time.Sleep(500 * time.Microsecond) // [推演：模拟延迟] 缓存查询
	// 模拟缓存 miss，回源 DB
	time.Sleep(2 * time.Millisecond) // [推演：模拟延迟] DB 查询
	resp := Response{
		Status: "ok",
		Data: map[string]string{
			"layer":  "2",
			"desc":   "cache-miss-then-db",
			"cached": "false",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func layer3Handler(w http.ResponseWriter, r *http.Request) {
	// Layer 3: +模拟外部HTTP调用
	// 缓存 miss -> DB 查询 -> 调用外部服务
	time.Sleep(500 * time.Microsecond) // [推演：模拟延迟] 缓存查询
	time.Sleep(2 * time.Millisecond)    // [推演：模拟延迟] DB 查询

	// 调用辅助 HTTP 服务器
	resp, err := http.Get("http://127.0.0.1:18091/external")
	if err != nil {
		http.Error(w, "external call failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	result := Response{
		Status: "ok",
		Data: map[string]string{
			"layer":      "3",
			"desc":       "cache-db-external",
			"ext-result": string(body),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func layer4Handler(w http.ResponseWriter, r *http.Request) {
	// Layer 4: +业务逻辑
	// 缓存 miss -> DB 查询 -> 外部调用 -> 字符串处理+排序
	time.Sleep(500 * time.Microsecond) // [推演：模拟延迟] 缓存查询
	time.Sleep(2 * time.Millisecond)    // [推演：模拟延迟] DB 查询

	// 调用辅助 HTTP 服务器
	resp, err := http.Get("http://127.0.0.1:18091/external")
	if err != nil {
		http.Error(w, "external call failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// [推演：模拟延迟] 业务逻辑：字符串处理 + 排序（约 1ms）
	items := make([]string, 0, 500)
	for i := 0; i < 500; i++ {
		items = append(items, fmt.Sprintf("item-%d-%s", i, string(body)))
	}
	sort.Strings(items)
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(item)
		sb.WriteString(";")
	}
	_ = sb.String()

	result := Response{
		Status: "ok",
		Data: map[string]string{
			"layer":           "4",
			"desc":            "full-stack",
			"items-processed": fmt.Sprintf("%d", len(items)),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ============================================================
// 压测引擎
// ============================================================

type BenchResult struct {
	Layer        string
	Requests     int
	Concurrency  int
	QPS          float64
	P50          time.Duration
	P90          time.Duration
	P99          time.Duration
	Min          time.Duration
	Max          time.Duration
	HeapAllocMB  float64
	TotalAllocMB float64
}

func benchLayer(port, path string, requests, concurrency int) *BenchResult {
	url := fmt.Sprintf("http://127.0.0.1:%s%s", port, path)

	// 采集压测前的内存信息
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	latencies := make([]time.Duration, requests)
	var wg sync.WaitGroup
	var idx atomic.Int64

	// 使用带缓冲的 channel 控制并发
	sem := make(chan struct{}, concurrency)

	start := time.Now()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			reqStart := time.Now()
			resp, err := http.Get(url)
			if err != nil {
				latencies[idx.Add(1)-1] = 0
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
			latencies[idx.Add(1)-1] = time.Since(reqStart)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	// 采集压测后的内存信息
	runtime.ReadMemStats(&mAfter)

	// 过滤掉失败的请求（latency = 0）
	validLatencies := make([]time.Duration, 0, len(latencies))
	for _, l := range latencies {
		if l > 0 {
			validLatencies = append(validLatencies, l)
		}
	}
	sort.Slice(validLatencies, func(i, j int) bool {
		return validLatencies[i] < validLatencies[j]
	})

	result := &BenchResult{
		Requests:     len(validLatencies),
		Concurrency:  concurrency,
		QPS:          float64(len(validLatencies)) / elapsed.Seconds(),
		HeapAllocMB:  float64(mAfter.HeapAlloc) / 1024 / 1024,
		TotalAllocMB: float64(mAfter.TotalAlloc-mBefore.TotalAlloc) / 1024 / 1024,
	}

	if len(validLatencies) > 0 {
		result.Min = validLatencies[0]
		result.Max = validLatencies[len(validLatencies)-1]
		result.P50 = validLatencies[int(float64(len(validLatencies))*0.50)]
		result.P90 = validLatencies[int(float64(len(validLatencies))*0.90)]
		result.P99 = validLatencies[int(float64(len(validLatencies))*0.99)]
	}

	return result
}

// ============================================================
// 辅助函数
// ============================================================

func fmtDur(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Microseconds()))
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
}

func fmtFloat(f float64, unit string) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "N/A"
	}
	return fmt.Sprintf("%.1f%s", f, unit)
}

// ============================================================
// main
// ============================================================

func main() {
	fmt.Println("=== Go HTTP 单机天花板实验 ===")
	fmt.Println("[实测 Go 1.26.2 darwin/arm64]")
	fmt.Println()

	// 启动辅助 HTTP 服务器
	fmt.Println("启动辅助 HTTP 服务器（模拟外部调用）...")
	startHelperServer("18091")

	// 启动主服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/layer0", layer0Handler)
	mux.HandleFunc("/layer1", layer1Handler)
	mux.HandleFunc("/layer2", layer2Handler)
	mux.HandleFunc("/layer3", layer3Handler)
	mux.HandleFunc("/layer4", layer4Handler)

	go http.ListenAndServe(":18090", mux)
	time.Sleep(300 * time.Millisecond) // 等待服务器启动

	fmt.Println("主服务器已启动 :18090, 辅助服务器 :18091")
	fmt.Println()

	// 预热：每层先发 100 个请求
	fmt.Println("预热中（每层 100 请求）...")
	for i := 0; i < 5; i++ {
		benchLayer("18090", fmt.Sprintf("/layer%d", i), 100, 10)
	}
	fmt.Println("预热完成")
	fmt.Println()

	// 正式压测
	const requests = 10000
	const concurrency = 100

	fmt.Printf("正式压测: %d 请求, %d 并发\n\n", requests, concurrency)

	layers := []struct {
		name string
		path string
	}{
		{"Layer 0: 纯JSON序列化", "/layer0"},
		{"Layer 1: +DB查询", "/layer1"},
		{"Layer 2: +Redis缓存", "/layer2"},
		{"Layer 3: +外部HTTP调用", "/layer3"},
		{"Layer 4: +业务逻辑", "/layer4"},
	}

	results := make([]*BenchResult, 0, len(layers))

	for _, l := range layers {
		fmt.Printf("压测 %s ...\n", l.name)
		r := benchLayer("18090", l.path, requests, concurrency)
		results = append(results, r)
	}

	// ============================================================
	// 输出结果表格
	// ============================================================

	fmt.Println()
	fmt.Println("=== 实验结果 ===")
	fmt.Println()
	fmt.Printf("[实测 Go 1.26.2 darwin/arm64] | %d 并发 | %d 请求/层\n\n", concurrency, requests)
	fmt.Println("| 层级 | QPS | P50 | P90 | P99 | Min | Max | 堆内存(MB) | 总分配(MB) |")
	fmt.Println("|------|-----|-----|-----|-----|-----|-----|-----------|-----------|")

	layerNames := []string{
		"L0:纯JSON",
		"L1:+DB",
		"L2:+Cache",
		"L3:+HTTP",
		"L4:+Logic",
	}

	for i, r := range results {
		fmt.Printf("| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			layerNames[i],
			fmtFloat(r.QPS, ""),
			fmtDur(r.P50),
			fmtDur(r.P90),
			fmtDur(r.P99),
			fmtDur(r.Min),
			fmtDur(r.Max),
			fmtFloat(r.HeapAllocMB, ""),
			fmtFloat(r.TotalAllocMB, ""),
		)
	}

	// ============================================================
	// 衰减分析
	// ============================================================

	fmt.Println()
	fmt.Println("=== 衰减分析 ===")
	fmt.Println()

	if len(results) > 0 && results[0].QPS > 0 {
		baseQPS := results[0].QPS
		fmt.Println("| 层级 | 相对L0 QPS | QPS衰减率 | 延迟增长倍数(P50) | 延迟增长倍数(P99) |")
		fmt.Println("|------|-----------|----------|------------------|------------------|")
		for i, r := range results {
			qpsRatio := r.QPS / baseQPS * 100
			decayRate := (1 - r.QPS/baseQPS) * 100
			p50Multiplier := float64(r.P50) / float64(results[0].P50)
			p99Multiplier := float64(r.P99) / float64(results[0].P99)
			fmt.Printf("| %s | %.1f%% | %.1f%% | %.1fx | %.1fx |\n",
				layerNames[i], qpsRatio, decayRate, p50Multiplier, p99Multiplier)
		}
	}

	// ============================================================
	// GC 信息
	// ============================================================

	fmt.Println()
	fmt.Println("=== GC 停顿信息 ===")
	fmt.Println()
	fmt.Println("注：runtime.ReadMemStats 本身会触发 STW，此数据仅作参考")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("GCSys: %.1f MB | HeapSys: %.1f MB | NumGC: %d | GCCPUFraction: %.6f\n",
		float64(m.GCSys)/1024/1024,
		float64(m.HeapSys)/1024/1024,
		m.NumGC,
		m.GCCPUFraction,
	)

	// 采集最近 GC 停顿
	if m.NumGC > 0 {
		fmt.Println("\n最近 GC 停顿（从 MemStats PauseNs 采集）:")
		count := 0
		for i := int(m.NumGC) - 1; i >= 0 && count < 10; i-- {
			fmt.Printf("  GC #%d: %s\n", i, time.Duration(m.PauseNs[i%256]))
			count++
		}
	}

	fmt.Println()
	fmt.Println("=== 实验完成 ===")
}
