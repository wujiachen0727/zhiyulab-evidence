// context-timeout-cascade: 3层HTTP调用链超时策略对比实验
// [实测 Go 1.26.2 darwin/arm64]
// [推演：模拟处理延迟]
//
// 验证：3种超时策略下 P99 延迟和级联失败率差异，以及 99.9%^N 可靠性衰减
//
// 设计思路：
// - DataLayer: 正常 5ms, 慢请求 30ms (20%), 极慢 80ms (5%)
// - Service: 2ms + 调 DataLayer
// - Gateway: 1ms + 调 Service
// - 最小总延迟: 1+2+5=8ms, 最大: 1+2+80=83ms
// - 固定超时 15ms → 慢请求全部超时，级联失败严重
// - 递减超时 15/12/8ms → 各层超时独立，上游不浪费资源
// - 自适应超时 → 基于上游剩余时间，平衡超时控制与成功率

package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 服务基础设施
// ============================================================

func startServer(port int, handler http.Handler) (*http.Server, string) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error on :%d: %v\n", port, err)
		}
	}()
	return srv, url
}

func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			io.ReadAll(resp.Body)
			resp.Body.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func newClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     200,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{Transport: transport}
}

// ============================================================
// DataLayer: 最底层，模拟数据访问
// 正常 5ms, 20% 慢请求 30ms, 5% 极慢 80ms [推演：模拟处理延迟]
// ============================================================

func newDataLayerHandler() http.Handler {
	mux := http.NewServeMux()
	rng := rand.New(rand.NewSource(42))
	var mu sync.Mutex

	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		p := rng.Float64()
		mu.Unlock()

		var delay time.Duration
		switch {
		case p < 0.05: // 5% 极慢请求
			delay = 80 * time.Millisecond
		case p < 0.25: // 20% 慢请求
			delay = 30 * time.Millisecond
		default: // 75% 正常
			delay = 5 * time.Millisecond
		}
		time.Sleep(delay) // [推演：模拟处理延迟]
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data-ok")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	return mux
}

// ============================================================
// Service: 中间层，处理 2ms + 调用 DataLayer [推演：模拟处理延迟]
// ============================================================

func newServiceHandler(dataLayerURL string, strategy string) http.Handler {
	mux := http.NewServeMux()
	client := newClient()

	mux.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Millisecond) // [推演：模拟处理延迟]

		var ctx context.Context
		switch strategy {
		case "fixed":
			// 固定超时：每层一样，不考虑上游剩余时间
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.Context(), 15*time.Millisecond)
			defer cancel()
		case "decreasing":
			// 递减超时：越往下游越短
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.Context(), 12*time.Millisecond)
			defer cancel()
		case "adaptive":
			// 自适应超时：基于上游剩余时间动态计算
			if deadline, ok := r.Context().Deadline(); ok {
				remaining := time.Until(deadline)
				safeMargin := 3 * time.Millisecond // 安全余量：本层处理时间
				if remaining > safeMargin {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(r.Context(), remaining-safeMargin)
					defer cancel()
				} else {
					// 剩余时间不足以安全调用下游，快速失败
					w.WriteHeader(http.StatusGatewayTimeout)
					fmt.Fprint(w, "service-no-margin")
					return
				}
			} else {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(r.Context(), 15*time.Millisecond)
				defer cancel()
			}
		default:
			ctx = r.Context()
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", dataLayerURL+"/data", nil)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprint(w, "service-upstream-timeout")
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(resp.StatusCode)
			w.Write(body)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	return mux
}

// ============================================================
// Gateway: 网关层，处理 1ms + 调用 Service [推演：模拟处理延迟]
// ============================================================

func newGatewayHandler(serviceURL string, strategy string) http.Handler {
	mux := http.NewServeMux()
	client := newClient()

	mux.HandleFunc("/gateway", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // [推演：模拟处理延迟]

		var ctx context.Context
		switch strategy {
		case "fixed":
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.Context(), 15*time.Millisecond)
			defer cancel()
		case "decreasing":
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.Context(), 15*time.Millisecond)
			defer cancel()
		case "adaptive":
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.Context(), 15*time.Millisecond)
			defer cancel()
		default:
			ctx = r.Context()
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", serviceURL+"/service", nil)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprint(w, "gateway-timeout")
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(resp.StatusCode)
			w.Write(body)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	return mux
}

// ============================================================
// 压测引擎
// ============================================================

type BenchResult struct {
	Strategy          string
	Latencies         []time.Duration
	SuccessCount      int64
	TimeoutCount      int64
	CascadeFailCount  int64 // Gateway 超时导致整链失败
	UpstreamFailCount int64 // 上游超时传播（Service/DataLayer 超时）
	TotalRequests     int
}

func bench(gatewayURL string, requests int, concurrency int, strategy string) *BenchResult {
	result := &BenchResult{
		Strategy:      strategy,
		TotalRequests: requests,
	}

	var latencyMu sync.Mutex
	latencies := make([]time.Duration, 0, requests)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	var successCount int64
	var timeoutCount int64
	var cascadeFailCount int64
	var upstreamFailCount int64

	client := newClient()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			req, _ := http.NewRequestWithContext(ctx, "GET", gatewayURL+"/gateway", nil)
			resp, err := client.Do(req)
			elapsed := time.Since(start)

			if err != nil {
				atomic.AddInt64(&timeoutCount, 1)
				atomic.AddInt64(&cascadeFailCount, 1)
			} else {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				bodyStr := string(body)
				if resp.StatusCode == http.StatusOK && bodyStr == "data-ok" {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&timeoutCount, 1)
					if bodyStr == "gateway-timeout" {
						atomic.AddInt64(&cascadeFailCount, 1)
					} else {
						atomic.AddInt64(&upstreamFailCount, 1)
					}
				}
			}

			latencyMu.Lock()
			latencies = append(latencies, elapsed)
			latencyMu.Unlock()
		}()
	}

	wg.Wait()

	result.Latencies = latencies
	result.SuccessCount = successCount
	result.TimeoutCount = timeoutCount
	result.CascadeFailCount = cascadeFailCount
	result.UpstreamFailCount = upstreamFailCount

	return result
}

func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Ceil(float64(len(sorted))*p/100.0)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func printResult(r *BenchResult) {
	p50 := percentile(r.Latencies, 50)
	p90 := percentile(r.Latencies, 90)
	p99 := percentile(r.Latencies, 99)

	successRate := float64(r.SuccessCount) / float64(r.TotalRequests) * 100
	timeoutRate := float64(r.TimeoutCount) / float64(r.TotalRequests) * 100
	cascadeRate := float64(r.CascadeFailCount) / float64(r.TotalRequests) * 100
	upstreamRate := float64(r.UpstreamFailCount) / float64(r.TotalRequests) * 100

	fmt.Printf("\n========== 策略: %s ==========\n", r.Strategy)
	fmt.Printf("总请求数: %d\n", r.TotalRequests)
	fmt.Printf("成功: %d | 超时: %d | 级联失败: %d | 上游失败: %d\n",
		r.SuccessCount, r.TimeoutCount, r.CascadeFailCount, r.UpstreamFailCount)
	fmt.Printf("P50: %v | P90: %v | P99: %v\n", p50, p90, p99)
	fmt.Printf("成功率: %.2f%% | 超时错误率: %.2f%% | 级联失败率: %.2f%% | 上游失败率: %.2f%%\n",
		successRate, timeoutRate, cascadeRate, upstreamRate)
	fmt.Printf("================================\n\n")
}

// ============================================================
// 主流程
// ============================================================

func runStrategy(strategy string, dataPort, svcPort, gwPort int) *BenchResult {
	dataLayerURL := fmt.Sprintf("http://127.0.0.1:%d", dataPort)
	serviceURL := fmt.Sprintf("http://127.0.0.1:%d", svcPort)
	gatewayURL := fmt.Sprintf("http://127.0.0.1:%d", gwPort)

	dataSrv, _ := startServer(dataPort, newDataLayerHandler())
	if !waitForServer(dataLayerURL+"/health", 5*time.Second) {
		fmt.Fprintf(os.Stderr, "DataLayer 启动失败\n")
		os.Exit(1)
	}

	svcSrv, _ := startServer(svcPort, newServiceHandler(dataLayerURL, strategy))
	if !waitForServer(serviceURL+"/health", 5*time.Second) {
		fmt.Fprintf(os.Stderr, "Service 启动失败\n")
		os.Exit(1)
	}

	gwSrv, _ := startServer(gwPort, newGatewayHandler(serviceURL, strategy))
	if !waitForServer(gatewayURL+"/health", 5*time.Second) {
		fmt.Fprintf(os.Stderr, "Gateway 启动失败\n")
		os.Exit(1)
	}

	// 预热
	warmupClient := newClient()
	for i := 0; i < 100; i++ {
		resp, err := warmupClient.Get(gatewayURL + "/gateway")
		if err == nil {
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
		time.Sleep(1 * time.Millisecond)
	}

	// 压测
	result := bench(gatewayURL, 20000, 100, strategy)

	// 关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	gwSrv.Shutdown(shutdownCtx)
	svcSrv.Shutdown(shutdownCtx)
	dataSrv.Shutdown(shutdownCtx)
	time.Sleep(1 * time.Second)

	return result
}

func main() {
	fmt.Println("=== 3层HTTP调用链超时策略对比实验 ===")
	fmt.Println("[实测 Go 1.26.2 darwin/arm64]")
	fmt.Println("[推演：模拟处理延迟]")
	fmt.Println("调用链: Gateway(1ms) → Service(2ms) → DataLayer(5ms/30ms/80ms)")
	fmt.Println("DataLayer 延迟分布: 75% 正常(5ms), 20% 慢(30ms), 5% 极慢(80ms)")
	fmt.Println("")
	fmt.Println("超时配置:")
	fmt.Println("  fixed:     Gateway 15ms, Service 15ms, DataLayer 无")
	fmt.Println("  decreasing: Gateway 15ms, Service 12ms, DataLayer 无")
	fmt.Println("  adaptive:  Gateway 15ms, Service 动态(上游剩余-3ms)")
	fmt.Println("压测参数: 100并发, 20000请求")
	fmt.Println("")

	strategies := []struct {
		name     string
		dataPort int
		svcPort  int
		gwPort   int
	}{
		{"fixed", 18088, 18089, 18090},
		{"decreasing", 18091, 18092, 18093},
		{"adaptive", 18094, 18095, 18096},
	}

	var results []*BenchResult
	for _, s := range strategies {
		fmt.Printf(">>> 运行策略: %s (ports %d/%d/%d)...\n", s.name, s.dataPort, s.svcPort, s.gwPort)
		r := runStrategy(s.name, s.dataPort, s.svcPort, s.gwPort)
		results = append(results, r)
		printResult(r)
	}

	// 汇总
	fmt.Println("=== 汇总 ===")
	fmt.Println("| 策略 | P50 | P90 | P99 | 成功率 | 超时错误率 | 级联失败率 | 上游失败率 |")
	fmt.Println("|------|-----|-----|-----|--------|-----------|-----------|-----------|")
	for _, r := range results {
		p50 := percentile(r.Latencies, 50)
		p90 := percentile(r.Latencies, 90)
		p99 := percentile(r.Latencies, 99)
		successRate := float64(r.SuccessCount) / float64(r.TotalRequests) * 100
		timeoutRate := float64(r.TimeoutCount) / float64(r.TotalRequests) * 100
		cascadeRate := float64(r.CascadeFailCount) / float64(r.TotalRequests) * 100
		upstreamRate := float64(r.UpstreamFailCount) / float64(r.TotalRequests) * 100
		fmt.Printf("| %s | %v | %v | %v | %.2f%% | %.2f%% | %.2f%% | %.2f%% |\n",
			r.Strategy, p50, p90, p99, successRate, timeoutRate, cascadeRate, upstreamRate)
	}
}
