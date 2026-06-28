package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============ 病因层：两种锁实现 ============

// 全局单锁——病因层（方案 A/B 用）
var globalMutex sync.Mutex

// 分片锁——病因已修复（方案 C 用）
const shardCount = 16
var shardedMutex [shardCount]sync.Mutex

func shardedLock(key uint64) {
	idx := key % shardCount
	shardedMutex[idx].Lock()
}

func shardedUnlock(key uint64) {
	idx := key % shardCount
	shardedMutex[idx].Unlock()
}

var requestCount uint64

// ============ 服务端 ============

// useShardedLock 控制使用哪种锁
var useShardedLock bool

func handler(w http.ResponseWriter, r *http.Request) {
	key := atomic.AddUint64(&requestCount, 1)

	if useShardedLock {
		shardedLock(key)
		time.Sleep(time.Duration(1+rand.Intn(2)) * time.Millisecond)
		shardedUnlock(key)
	} else {
		globalMutex.Lock()
		time.Sleep(time.Duration(1+rand.Intn(2)) * time.Millisecond)
		globalMutex.Unlock()
	}

	// 临界区外的慢操作（长尾）
	latency := 0
	if rand.Float64() < 0.10 {
		latency = 200 + rand.Intn(300)
	} else {
		latency = 10 + rand.Intn(40)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)

	fmt.Fprintf(w, "ok\n")
}

// ============ hedging client ============

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

// ============ 基准测试 ============

type metrics struct {
	latencies []time.Duration
	hedged    int32
}

func runScenario(useHedging bool, totalReq, conc int) *metrics {
	m := &metrics{latencies: make([]time.Duration, 0, totalReq)}
	client := &http.Client{Timeout: 5 * time.Second}
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
			resp, err := client.Get("http://127.0.0.1:8200/work")
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

func fetchMutexTotalWait() uint64 {
	resp, err := http.Get("http://127.0.0.1:8200/debug/pprof/mutex")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	_ = data
	return 0
}

var _ = fetchMutexTotalWait

func main() {
	runtime.SetMutexProfileFraction(100)

	mux := http.NewServeMux()
	mux.HandleFunc("/work", handler)
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	server := &http.Server{Addr: ":8200", Handler: mux}
	go func() {
		log.Println("server on :8200")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	// 预热
	for i := 0; i < 20; i++ {
		resp, _ := http.Get("http://127.0.0.1:8200/work")
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	time.Sleep(1 * time.Second)

	const totalReq = 2000
	const conc = 50

	// === 方案 A：无优化基线（单锁 + 无 hedging）===
	useShardedLock = false
	atomic.StoreUint64(&requestCount, 0)
	mA := runScenario(false, totalReq, conc)
	time.Sleep(2 * time.Second)
	countA := atomic.LoadUint64(&requestCount)
	// 保存 mutex profile
	resp, _ := http.Get("http://127.0.0.1:8200/debug/pprof/mutex")
	if resp != nil {
		out, _ := os.Create("mutex_profile_A.pb.gz")
		io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
	}

	fmt.Println("=== 方案 A：无优化基线（单锁 + 无 hedging）===")
	fmt.Printf("P50=%v P95=%v P99=%v (n=%d, 服务端 %d)\n",
		percentile(mA.latencies, 0.50),
		percentile(mA.latencies, 0.95),
		percentile(mA.latencies, 0.99),
		len(mA.latencies), countA)

	time.Sleep(3 * time.Second)

	// === 方案 B：仅 hedging（单锁 + hedging）——症状治疗 ===
	useShardedLock = false
	atomic.StoreUint64(&requestCount, 0)
	mB := runScenario(true, totalReq, conc)
	time.Sleep(2 * time.Second)
	countB := atomic.LoadUint64(&requestCount)
	resp, _ = http.Get("http://127.0.0.1:8200/debug/pprof/mutex")
	if resp != nil {
		out, _ := os.Create("mutex_profile_B.pb.gz")
		io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
	}

	fmt.Println("\n=== 方案 B：仅 hedging（单锁 + hedging）===")
	fmt.Printf("P50=%v P95=%v P99=%v (n=%d, 服务端 %d, hedging %d)\n",
		percentile(mB.latencies, 0.50),
		percentile(mB.latencies, 0.95),
		percentile(mB.latencies, 0.99),
		len(mB.latencies), countB, mB.hedged)

	time.Sleep(3 * time.Second)

	// === 方案 C：修复锁竞争 + hedging（分片锁 + hedging）——病因治疗 + 症状兜底 ===
	useShardedLock = true
	atomic.StoreUint64(&requestCount, 0)
	mC := runScenario(true, totalReq, conc)
	time.Sleep(2 * time.Second)
	countC := atomic.LoadUint64(&requestCount)
	resp, _ = http.Get("http://127.0.0.1:8200/debug/pprof/mutex")
	if resp != nil {
		out, _ := os.Create("mutex_profile_C.pb.gz")
		io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
	}

	fmt.Println("\n=== 方案 C：修复锁竞争 + hedging（分片锁 + hedging）===")
	fmt.Printf("P50=%v P95=%v P99=%v (n=%d, 服务端 %d, hedging %d)\n",
		percentile(mC.latencies, 0.50),
		percentile(mC.latencies, 0.95),
		percentile(mC.latencies, 0.99),
		len(mC.latencies), countC, mC.hedged)

	// === 对比 ===
	fmt.Println("\n=== 三层对比 ===")
	fmt.Printf("%-28s %12s %12s %12s\n", "方案", "P50", "P95", "P99")
	fmt.Printf("%-28s %12v %12v %12v\n", "A 无优化", percentile(mA.latencies, 0.50), percentile(mA.latencies, 0.95), percentile(mA.latencies, 0.99))
	fmt.Printf("%-28s %12v %12v %12v\n", "B 仅 hedging", percentile(mB.latencies, 0.50), percentile(mB.latencies, 0.95), percentile(mB.latencies, 0.99))
	fmt.Printf("%-28s %12v %12v %12v\n", "C 修复+hedging", percentile(mC.latencies, 0.50), percentile(mC.latencies, 0.95), percentile(mC.latencies, 0.99))
	fmt.Printf("\n%-28s %12s %12s\n", "方案", "服务端请求量")
	fmt.Printf("%-28s %12d %12s\n", "A 无优化", countA, "")
	fmt.Printf("%-28s %12d %12s\n", "B 仅 hedging", countB, "")
	fmt.Printf("%-28s %12d %12s\n", "C 修复+hedging", countC, "")

	server.Shutdown(context.Background())
}
