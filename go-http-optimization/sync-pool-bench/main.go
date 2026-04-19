package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 模拟业务响应结构体：5-10个字段 + 嵌套对象
type NestedDetail struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Score     float64 `json:"score"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
}

type ApiResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []NestedDetail `json:"data"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	Size    int            `json:"size"`
	Meta    struct {
		RequestID string `json:"request_id"`
		Timestamp int64  `json:"timestamp"`
	} `json:"meta"`
}

func makeResponse() *ApiResponse {
	resp := &ApiResponse{
		Code:    200,
		Message: "success",
		Total:   100,
		Page:    1,
		Size:    10,
	}
	resp.Meta.RequestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
	resp.Meta.Timestamp = time.Now().Unix()
	for i := 0; i < 10; i++ {
		resp.Data = append(resp.Data, NestedDetail{
			ID:        i + 1,
			Name:      fmt.Sprintf("item-%d", i),
			Score:     float64(i) * 1.23,
			IsActive:  i%2 == 0,
			CreatedAt: time.Now().Format(time.RFC3339),
		})
	}
	return resp
}

// ============================================================
// 无 sync.Pool 版本
// ============================================================

func startNoPoolServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		resp := makeResponse()
		buf := &bytes.Buffer{}
		json.NewEncoder(buf).Encode(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	})
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// ============================================================
// 有 sync.Pool 版本
// ============================================================

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 2048))
	},
}

func startWithPoolServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		resp := makeResponse()
		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		json.NewEncoder(buf).Encode(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
		bufPool.Put(buf)
	})
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// ============================================================
// 客户端压测
// ============================================================

func runBench(port string, requests, concurrency int) {
	url := fmt.Sprintf("http://127.0.0.1:%s/api/data", port)

	// 配置连接池：复用 TCP 连接，避免端口耗尽
	transport := &http.Transport{
		MaxIdleConnsPerHost: concurrency,
		MaxConnsPerHost:     concurrency,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// 记录 GC 前快照
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)
	gcBefore := memBefore.NumGC

	type result struct {
		lat time.Duration
		ok  bool
	}

	jobs := make(chan int, concurrency*2)
	results := make(chan result, concurrency*2)

	// 启动固定数量的 worker
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				reqStart := time.Now()
				resp, err := client.Get(url)
				if err != nil {
					results <- result{lat: 0, ok: false}
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				lat := time.Since(reqStart)
				results <- result{lat: lat, ok: true}
			}
		}()
	}

	// 后台发送任务
	go func() {
		for i := 0; i < requests; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	startTime := time.Now()

	// 等待所有 worker 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var latencies []time.Duration
	var errCounter int64
	for r := range results {
		if r.ok {
			latencies = append(latencies, r.lat)
		} else {
			atomic.AddInt64(&errCounter, 1)
		}
	}
	duration := time.Since(startTime)

	// 记录 GC 后快照
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	gcAfter := memAfter.NumGC

	// 从 PauseNs 环形缓冲区提取 GC 停顿
	var gcPauses []time.Duration
	for i := gcBefore; i < gcAfter; i++ {
		idx := i % 256
		gcPauses = append(gcPauses, time.Duration(memAfter.PauseNs[idx]))
	}

	qps := float64(len(latencies)) / duration.Seconds()

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := percentile(latencies, 50)
	p90 := percentile(latencies, 90)
	p99 := percentile(latencies, 99)

	var gcPauseTotal time.Duration
	for _, p := range gcPauses {
		gcPauseTotal += p
	}

	fmt.Fprintf(os.Stderr, "BENCH_RESULT_START\n")
	fmt.Fprintf(os.Stderr, "completed_requests: %d\n", len(latencies))
	fmt.Fprintf(os.Stderr, "error_requests: %d\n", errCounter)
	fmt.Fprintf(os.Stderr, "duration_ms: %d\n", duration.Milliseconds())
	fmt.Fprintf(os.Stderr, "qps: %.0f\n", qps)
	fmt.Fprintf(os.Stderr, "p50_ns: %d\n", p50.Nanoseconds())
	fmt.Fprintf(os.Stderr, "p90_ns: %d\n", p90.Nanoseconds())
	fmt.Fprintf(os.Stderr, "p99_ns: %d\n", p99.Nanoseconds())
	fmt.Fprintf(os.Stderr, "gc_count: %d\n", gcAfter-gcBefore)
	fmt.Fprintf(os.Stderr, "gc_pause_total_ns: %d\n", gcPauseTotal.Nanoseconds())
	fmt.Fprintf(os.Stderr, "gc_pause_count: %d\n", len(gcPauses))
	if len(gcPauses) > 0 {
		sort.Slice(gcPauses, func(i, j int) bool { return gcPauses[i] < gcPauses[j] })
		fmt.Fprintf(os.Stderr, "gc_pause_avg_ns: %d\n", (gcPauseTotal/time.Duration(len(gcPauses))).Nanoseconds())
		fmt.Fprintf(os.Stderr, "gc_pause_p50_ns: %d\n", gcPauses[len(gcPauses)*50/100].Nanoseconds())
		fmt.Fprintf(os.Stderr, "gc_pause_p90_ns: %d\n", gcPauses[len(gcPauses)*90/100].Nanoseconds())
		fmt.Fprintf(os.Stderr, "gc_pause_p99_ns: %d\n", gcPauses[len(gcPauses)*99/100].Nanoseconds())
		fmt.Fprintf(os.Stderr, "gc_pause_max_ns: %d\n", gcPauses[len(gcPauses)-1].Nanoseconds())
	}
	fmt.Fprintf(os.Stderr, "mem_mallocs_delta: %d\n", memAfter.Mallocs-memBefore.Mallocs)
	fmt.Fprintf(os.Stderr, "mem_frees_delta: %d\n", memAfter.Frees-memBefore.Frees)
	fmt.Fprintf(os.Stderr, "mem_total_alloc_delta_mb: %d\n", (memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
	fmt.Fprintf(os.Stderr, "mem_heap_alloc_kb: %d\n", memAfter.HeapAlloc/1024)
	fmt.Fprintf(os.Stderr, "BENCH_RESULT_END\n")
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func main() {
	mode := flag.String("mode", "", "运行模式: nopool-server, withpool-server, bench")
	port := flag.String("port", "", "端口")
	requests := flag.Int("requests", 50000, "总请求数")
	concurrency := flag.Int("concurrency", 100, "并发数")
	flag.Parse()

	switch *mode {
	case "nopool-server":
		startNoPoolServer(*port)
	case "withpool-server":
		startWithPoolServer(*port)
	case "bench":
		runBench(*port, *requests, *concurrency)
	default:
		fmt.Fprintf(os.Stderr, "用法: %s -mode <nopool-server|withpool-server|bench> -port <port>\n", os.Args[0])
		os.Exit(1)
	}
}
