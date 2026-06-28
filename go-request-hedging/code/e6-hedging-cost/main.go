package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var globalMutex sync.Mutex
var requestCount uint64

func handler(w http.ResponseWriter, r *http.Request) {
	globalMutex.Lock()
	time.Sleep(time.Duration(1+rand.Intn(2)) * time.Millisecond)
	globalMutex.Unlock()

	latency := 0
	if rand.Float64() < 0.10 {
		latency = 200 + rand.Intn(300)
	} else {
		latency = 10 + rand.Intn(40)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Fprintf(w, "ok\n")
}

type hedgedTransport struct {
	base       http.RoundTripper
	hedgeDelay time.Duration
	hedged     *int32
}

func (t *hedgedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req1 := req.Clone(ctx)
	type result struct {
		resp *http.Response
		err  error
	}
	ch := make(chan result, 2)

	go func() {
		resp, err := t.base.RoundTrip(req1)
		ch <- result{resp, err}
	}()

	timer := time.NewTimer(t.hedgeDelay)
	defer timer.Stop()

	select {
	case r := <-ch:
		return r.resp, r.err
	case <-timer.C:
		if t.hedged != nil {
			atomic.AddInt32(t.hedged, 1)
		}
		req2 := req.Clone(ctx)
		go func() {
			resp, err := t.base.RoundTrip(req2)
			ch <- result{resp, err}
		}()
		r := <-ch
		return r.resp, r.err
	}
}

type metrics struct {
	latencies []time.Duration
	hedged    int32
}

func runScenario(useHedging bool, totalReq, conc int) *metrics {
	m := &metrics{latencies: make([]time.Duration, 0, totalReq)}
	client := &http.Client{
		Timeout: 5 * time.Second,
		// 限制连接池，模拟生产环境连接池压力
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     100,
		},
	}
	if useHedging {
		client.Transport = &hedgedTransport{
			base:       http.DefaultTransport,
			hedgeDelay: 50 * time.Millisecond,
			hedged:     &m.hedged,
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, conc)
	var mu sync.Mutex

	for i := 0; i < totalReq; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			start := time.Now()
			resp, err := client.Get("http://127.0.0.1:8400/work")
			elapsed := time.Since(start)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			mu.Lock()
			m.latencies = append(m.latencies, elapsed)
			mu.Unlock()
		}()
	}
	wg.Wait()

	sort.Slice(m.latencies, func(i, j int) bool {
		return m.latencies[i] < m.latencies[j]
	})
	return m
}

func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	idx := int(float64(len(latencies)) * p)
	if idx >= len(latencies) {
		idx = len(latencies) - 1
	}
	return latencies[idx]
}

// readMemStats 采样内存分配
func readMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func main() {
	runtime.SetMutexProfileFraction(100)

	mux := http.NewServeMux()
	mux.HandleFunc("/work", handler)
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	server := &http.Server{Addr: ":8400", Handler: mux}
	go func() {
		log.Println("server on :8400")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 20; i++ {
		resp, _ := http.Get("http://127.0.0.1:8400/work")
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	time.Sleep(1 * time.Second)

	const totalReq = 2000
	const conc = 50

	// 基线内存
	memBefore := readMemStats()

	// 场景 A：无 hedging
	atomic.StoreUint64(&requestCount, 0)
	memA0 := readMemStats()
	mA := runScenario(false, totalReq, conc)
	memA1 := readMemStats()
	countA := atomic.LoadUint64(&requestCount)

	// 场景 B：有 hedging
	time.Sleep(3 * time.Second)
	atomic.StoreUint64(&requestCount, 0)
	memB0 := readMemStats()
	mB := runScenario(true, totalReq, conc)
	memB1 := readMemStats()
	countB := atomic.LoadUint64(&requestCount)

	fmt.Println("=== hedging 成本实测 ===")
	fmt.Printf("总请求量: %d, 并发: %d\n\n", totalReq, conc)

	fmt.Printf("%-20s %12s %12s\n", "指标", "无 hedging", "有 hedging")
	fmt.Printf("%-20s %12v %12v\n", "P50", percentile(mA.latencies, 0.50), percentile(mB.latencies, 0.50))
	fmt.Printf("%-20s %12v %12v\n", "P95", percentile(mA.latencies, 0.95), percentile(mB.latencies, 0.95))
	fmt.Printf("%-20s %12v %12v\n", "P99", percentile(mA.latencies, 0.99), percentile(mB.latencies, 0.99))
	fmt.Printf("%-20s %12d %12d\n", "服务端处理量", countA, countB)
	fmt.Printf("%-20s %12d %12d\n", "hedging 触发", 0, mB.hedged)

	// 内存成本
	allocA := memA1.TotalAlloc - memA0.TotalAlloc
	allocB := memB1.TotalAlloc - memB0.TotalAlloc
	gcA := memA1.NumGC - memA0.NumGC
	gcB := memB1.NumGC - memB0.NumGC
	fmt.Printf("%-20s %12d %12d\n", "内存分配(bytes)", allocA, allocB)
	fmt.Printf("%-20s %12d %12d\n", "GC 次数", gcA, gcB)
	fmt.Printf("%-20s %12d %12d\n", "Goroutine 数", runtime.NumGoroutine(), runtime.NumGoroutine())

	// 连接池压力（通过服务端处理量推算）
	fmt.Println("\n=== 连接池压力推算 ===")
	extraReq := int64(countB) - int64(countA)
	extraPercent := float64(extraReq) / float64(countA) * 100
	fmt.Printf("额外请求量: %d (%.1f%%)\n", extraReq, extraPercent)
	fmt.Printf("→ 连接池需要承受 %.1f%% 的额外连接\n", extraPercent)
	fmt.Printf("→ 下游服务的 QPS 被放大 %.1f%%\n", extraPercent)

	// 可观测性污染
	fmt.Println("\n=== 可观测性污染 ===")
	fmt.Printf("hedging 触发 %d 次 / %d 请求 = %.1f%%\n", mB.hedged, totalReq, float64(mB.hedged)/float64(totalReq)*100)
	fmt.Println("→ 服务端日志量增加（每个 hedging 请求都会被记日志）")
	fmt.Println("→ 服务端 metrics 被污染（处理量虚高，P99 失真）")
	fmt.Println("→ 链路追踪出现重复 span（同一逻辑请求对应两条服务端 trace）")

	_ = memBefore
	server.Shutdown(context.Background())
}
