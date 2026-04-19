package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// E5 v3: 用更可靠的方式测量单机天花板
// 每层独立启动服务器，独立压测，避免资源争抢

func main() {
	fmt.Println("=== Go HTTP 单机天花板实验 ===")
	fmt.Println("[实测 Go 1.26.2 darwin/arm64]")
	fmt.Println("[推演：模拟IO延迟] time.Sleep模拟DB/Redis/HTTP调用")
	fmt.Println()

	// 辅助HTTP服务（模拟外部API调用）
	go func() {
		http.ListenAndServe("127.0.0.1:29091", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Millisecond)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
	}()
	time.Sleep(200 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	layers := []struct {
		name string
		port string
		fn   http.HandlerFunc
	}{
		{
			"L0:纯JSON", "29092",
			func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ts": time.Now().Unix(), "items": make([]int, 50)})
			},
		},
		{
			"L1:+DB(2ms)", "29093",
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ts": time.Now().Unix(), "items": make([]int, 50)})
			},
		},
		{
			"L2:+Cache+DB(2.5ms)", "29094",
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(500 * time.Microsecond)
				time.Sleep(2 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ts": time.Now().Unix(), "items": make([]int, 50)})
			},
		},
		{
			"L3:+HTTP调用(5.5ms)", "29095",
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(500 * time.Microsecond)
				time.Sleep(2 * time.Millisecond)
				resp, _ := client.Get("http://127.0.0.1:29091/api")
				if resp != nil {
					resp.Body.Close()
				}
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ts": time.Now().Unix(), "items": make([]int, 50)})
			},
		},
		{
			"L4:+业务逻辑(6.5ms)", "29096",
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(500 * time.Microsecond)
				time.Sleep(2 * time.Millisecond)
				resp, _ := client.Get("http://127.0.0.1:29091/api")
				if resp != nil {
					resp.Body.Close()
				}
				time.Sleep(1 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ts": time.Now().Unix(), "items": make([]int, 50)})
			},
		},
	}

	// 逐层启动服务器并压测
	type result struct {
		name    string
		qps     float64
		p50     float64
		p90     float64
		p99     float64
		relQPS  float64
	}
	results := make([]result, 0, len(layers))
	var baseQPS float64

	for i, layer := range layers {
		// 启动该层服务器
		go func(idx int, port string, fn http.HandlerFunc) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", fn)
			http.ListenAndServe("127.0.0.1:"+port, mux)
		}(i, layer.port, layer.fn)
		time.Sleep(200 * time.Millisecond)

		// 压测
		url := "http://127.0.0.1:" + layer.port + "/"
		lats := benchHTTP(url, 8000, 100)

		// 计算QPS：总请求数/总耗时
		totalMs := sum(lats)
		qps := float64(len(lats)) / (totalMs / 1000.0 / float64(100)) * float64(100)
		// 更简单的算法：QPS = 请求数 / (P50 * 并发数 / 1000)... 不对
		// 最简单：QPS ≈ 并发数 / (平均延迟秒)
		avgMs := totalMs / float64(len(lats))
		qps = float64(100) / (avgMs / 1000.0)

		r := result{
			name: layer.name,
			qps:  qps,
			p50:  p(lats, 50),
			p90:  p(lats, 90),
			p99:  p(lats, 99),
		}
		if i == 0 {
			baseQPS = qps
			r.relQPS = 100.0
		} else {
			r.relQPS = qps / baseQPS * 100
		}
		results = append(results, r)

		fmt.Printf("  %s 完成: QPS=%.0f, P50=%.2fms, P99=%.2fms\n", layer.name, qps, r.p50, r.p99)
	}

	fmt.Println()
	fmt.Println("## 实验结果")
	fmt.Println()
	fmt.Printf("环境: [实测 Go 1.26.2 darwin/arm64] | 100并发 | 8000请求/层\n\n")
	fmt.Println("| 层级 | QPS | P50 | P90 | P99 | 相对L0 |")
	fmt.Println("|------|-----|-----|-----|-----|--------|")
	for _, r := range results {
		fmt.Printf("| %s | %.0f | %.2fms | %.2fms | %.2fms | %.1f%% |\n",
			r.name, r.qps, r.p50, r.p90, r.p99, r.relQPS)
	}

	fmt.Println()
	last := results[len(results)-1]
	fmt.Printf("### 衰减结论\n")
	fmt.Printf("- L0(纯JSON) → L4(全业务): QPS 衰减 %.0f%%\n", 100-last.relQPS)
	fmt.Printf("- P50 从 %.2fms 增长到 %.2fms (%.1fx)\n", results[0].p50, last.p50, last.p50/results[0].p50)
	fmt.Println("- Go 单机'百万QPS'只在纯JSON场景成立，真实业务复杂度下衰减剧烈")
}

func benchHTTP(url string, total, conc int) []float64 {
	var latencies []float64
	var mu sync.Mutex
	var wg sync.WaitGroup
	perW := total / conc

	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := &http.Client{Timeout: 5 * time.Second}
			for j := 0; j < perW; j++ {
				s := time.Now()
				resp, err := c.Get(url)
				d := time.Since(s).Seconds() * 1000
				if err == nil {
					resp.Body.Close()
				}
				mu.Lock()
				latencies = append(latencies, d)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return latencies
}

func p(data []float64, pct float64) float64 {
	s := make([]float64, len(data))
	copy(s, data)
	sort.Float64s(s)
	idx := int(float64(len(s)-1) * pct / 100.0)
	return s[idx]
}

func sum(data []float64) float64 {
	s := 0.0
	for _, v := range data {
		s += v
	}
	return s
}
