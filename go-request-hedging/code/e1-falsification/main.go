package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 全局 mutex 模拟生产环境热点资源的临界区
// 实验核心：hedging 降客户端 P99，但不应改善这个锁的竞争
var globalMutex sync.Mutex

// mutexWaitNs 累计 mutex 等待时间（采样近似）
// 通过 runtime/pprof mutex profile 读取
var requestCount uint64

func handler(w http.ResponseWriter, r *http.Request) {
	// 进入临界区——这是"病因层"：模拟热点资源的短临界区
	// 关键：mutex 只保护短操作（~1ms），不主导整体延迟
	// 这样 hedging 能绕过"个别请求慢"（长尾），但 mutex 等待时间不会因 hedging 改善
	globalMutex.Lock()
	// 模拟临界区内的短操作：1-3ms 的计数器更新或缓存写入
	time.Sleep(time.Duration(1+rand.Intn(2)) * time.Millisecond)
	atomic.AddUint64(&requestCount, 1)
	globalMutex.Unlock()

	// 模拟临界区外的慢操作：90% 概率 10-50ms，10% 概率 200-500ms（长尾）
	// 这部分是 hedging 能"绕过"的——发第二个请求可能命中快的节点
	latency := 0
	if rand.Float64() < 0.10 {
		latency = 200 + rand.Intn(300)
	} else {
		latency = 10 + rand.Intn(40)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)

	fmt.Fprintf(w, "ok\n")
}

// hedgedTransport 实现 hedging：第一个请求发出后，hedgeDelay 内未返回则发第二个，取先到的
type hedgedTransport struct {
	base       http.RoundTripper
	hedgeDelay time.Duration
	hedged     *int32 // 触发 hedging 的次数
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

	// 第一个请求
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
		// 触发 hedging
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

func runScenario(useHedging bool, totalReq, concurrency int) *metrics {
	m := &metrics{
		latencies: make([]time.Duration, 0, totalReq),
	}

	client := &http.Client{Timeout: 5 * time.Second}
	if useHedging {
		client.Transport = &hedgedTransport{
			base:       http.DefaultTransport,
			hedgeDelay: 50 * time.Millisecond, // P95 附近，正常请求 10-50ms，长尾 200-500ms
			hedged:     &m.hedged,
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex

	for i := 0; i < totalReq; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			resp, err := client.Get("http://127.0.0.1:8100/work")
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

// readMutexContention 从 runtime/pprof 读取 mutex profile，返回总等待时间（纳秒）
func readMutexContention() (totalWaitNs uint64, events int64) {
	// 通过 pprof 端点获取（服务端已注册 _ "net/http/pprof"）
	resp, err := http.Get("http://127.0.0.1:8100/debug/pprof/mutex")
	if err != nil {
		log.Printf("fetch mutex profile: %v", err)
		return 0, 0
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0
	}
	// 简化处理：profile 是 protobuf，我们无法在这里解析完整结构
	// 改为用 runtime.MemStats 类似的方式：直接读 profile 大小作为粗略活动度指示
	// 真正的解析在主实验中用 go tool pprof 完成
	_ = data
	return uint64(len(data)), 0
}

func main() {
	// 开启 mutex profile 采样（1/100）
	runtime.SetMutexProfileFraction(100)

	// 启动服务端
	mux := http.NewServeMux()
	mux.HandleFunc("/work", handler)
	// pprof 注册
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/mutex", http.DefaultServeMux.ServeHTTP)
	server := &http.Server{Addr: ":8100", Handler: mux}
	go func() {
		log.Println("server on :8100")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// 等服务端就绪
	time.Sleep(500 * time.Millisecond)

	// 预热
	for i := 0; i < 20; i++ {
		resp, _ := http.Get("http://127.0.0.1:8100/work")
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	time.Sleep(1 * time.Second)

	const totalReq = 2000
	const conc = 50

	// === 场景 A：无 hedging ===
	atomic.StoreUint64(&requestCount, 0)
	ma := runScenario(false, totalReq, conc)
	time.Sleep(2 * time.Second) // 等 mutex 采样
	profileA, _ := readMutexContention()
	countA := atomic.LoadUint64(&requestCount)

	fmt.Println("=== 场景 A：无 hedging ===")
	fmt.Printf("P50=%v P95=%v P99=%v (n=%d, 服务端处理 %d)\n",
		percentile(ma.latencies, 0.50),
		percentile(ma.latencies, 0.95),
		percentile(ma.latencies, 0.99),
		len(ma.latencies), countA)
	fmt.Printf("mutex profile 大小: %d bytes\n", profileA)

	// 保存 mutex profile
	resp, _ := http.Get("http://127.0.0.1:8100/debug/pprof/mutex")
	if resp != nil {
		out, _ := os.Create("mutex_profile_A.pb.gz")
		io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
	}

	// 等服务端恢复
	time.Sleep(3 * time.Second)

	// === 场景 B：有 hedging ===
	atomic.StoreUint64(&requestCount, 0)
	mb := runScenario(true, totalReq, conc)
	time.Sleep(2 * time.Second)
	profileB, _ := readMutexContention()
	countB := atomic.LoadUint64(&requestCount)

	fmt.Println("\n=== 场景 B：有 hedging ===")
	fmt.Printf("P50=%v P95=%v P99=%v (n=%d, 服务端处理 %d, hedging 触发 %d 次)\n",
		percentile(mb.latencies, 0.50),
		percentile(mb.latencies, 0.95),
		percentile(mb.latencies, 0.99),
		len(mb.latencies), countB, mb.hedged)
	fmt.Printf("mutex profile 大小: %d bytes\n", profileB)

	// 保存 mutex profile
	resp, _ = http.Get("http://127.0.0.1:8100/debug/pprof/mutex")
	if resp != nil {
		out, _ := os.Create("mutex_profile_B.pb.gz")
		io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
	}

	// === 对比 ===
	fmt.Println("\n=== 对比 ===")
	p99A := percentile(ma.latencies, 0.99)
	p99B := percentile(mb.latencies, 0.99)
	if p99A > 0 {
		drop := float64(p99A-p99B) / float64(p99A) * 100
		fmt.Printf("P99 下降: %.1f%% (从 %v 到 %v)\n", drop, p99A, p99B)
	}
	fmt.Printf("服务端处理请求量: %d → %d (hedging 导致 %d 额外请求)\n",
		countA, countB, int64(countB)-int64(countA))
	fmt.Printf("mutex profile 大小: %d → %d bytes\n", profileA, profileB)
	fmt.Println("\n[关键判定]")
	fmt.Println("如果 P99 明显下降但 mutex 等待时间相近 → 假设成立（hedging 是症状治疗）")
	fmt.Println("如果 mutex 等待时间也明显下降 → 假设被推翻（hedging 不只是症状治疗）")
	fmt.Println("→ 用 'go tool pprof -top mutex_profile_A.pb.gz' 和 'go tool pprof -top mutex_profile_B.pb.gz' 对比总等待时间")

	server.Shutdown(context.Background())
}
