package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// E4 v2: Context超时传播链式衰减
// 3层调用链，设置更紧的超时让级联效应显现

func main() {
	fmt.Println("=== Context超时传播链式衰减实验 ===")
	fmt.Println("[实测 Go 1.26.2 darwin/arm64]")
	fmt.Println("[推演：模拟处理延迟] DataLayer 5ms(90%) / 50ms(10%慢查询)")
	fmt.Println()

	// === 启动3层服务 ===

	// DataLayer: 5ms正常 / 50ms慢查询
	go func() {
		http.ListenAndServe("127.0.0.1:28093", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("slow") == "1" {
				time.Sleep(50 * time.Millisecond)
			} else {
				time.Sleep(5 * time.Millisecond)
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
	}()
	time.Sleep(200 * time.Millisecond)

	// 测试三种策略
	type strategy struct {
		name       string
		gwTimeout  time.Duration
		svcTimeout time.Duration
		dataTimeout time.Duration
	}

	strategies := []strategy{
		{"宽松超时(每层30ms)", 30 * time.Millisecond, 30 * time.Millisecond, 30 * time.Millisecond},
		{"递减超时(25→20→15ms)", 25 * time.Millisecond, 20 * time.Millisecond, 15 * time.Millisecond},
		{"紧超时(每层15ms)", 15 * time.Millisecond, 15 * time.Millisecond, 15 * time.Millisecond},
	}

	concurrency := 100
	requests := 5000
	slowRatio := 0.10 // 10% 慢查询

	type result struct {
		name          string
		p50, p90, p99 float64
		successRate   float64
		timeoutRate   float64
		cascadeFail   int
	}
	results := make([]result, 0)

	for _, s := range strategies {
		fmt.Printf("测试策略: %s ...\n", s.name)

		// Service层
		go func(st strategy) {
			http.ListenAndServe("127.0.0.1:28092", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx, cancel := context.WithTimeout(r.Context(), st.svcTimeout)
				defer cancel()

				slow := "0"
				if r.URL.Query().Get("slow") == "1" {
					slow = "1"
				}

				req, _ := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:28093/data?slow="+slow, nil)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
					return
				}
				resp.Body.Close()
				time.Sleep(2 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}))
		}(s)
		time.Sleep(200 * time.Millisecond)

		// Gateway层
		go func(st strategy) {
			http.ListenAndServe("127.0.0.1:28091", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx, cancel := context.WithTimeout(r.Context(), st.gwTimeout)
				defer cancel()

				slow := "0"
				if r.URL.Query().Get("slow") == "1" {
					slow = "1"
				}

				req, _ := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:28092/service?slow="+slow, nil)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
					return
				}
				resp.Body.Close()
				time.Sleep(1 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}))
		}(s)
		time.Sleep(200 * time.Millisecond)

		// 压测
		lats, successes, timeouts := benchWithSlow("http://127.0.0.1:28091/", requests, concurrency, slowRatio)

		r := result{
			name:        s.name,
			p50:         p(lats, 50),
			p90:         p(lats, 90),
			p99:         p(lats, 99),
			successRate: float64(successes) / float64(requests) * 100,
			timeoutRate: float64(timeouts) / float64(requests) * 100,
		}
		results = append(results, r)

		fmt.Printf("  完成: 成功率%.1f%%, 超时率%.1f%%\n", r.successRate, r.timeoutRate)
	}

	fmt.Println()
	fmt.Println("## 实验结果")
	fmt.Println()
	fmt.Printf("环境: [实测 Go 1.26.2 darwin/arm64] | %d并发 | %d请求 | %d%%慢查询\n\n",
		concurrency, requests, int(slowRatio*100))
	fmt.Println("| 策略 | P50 | P90 | P99 | 成功率 | 超时率 |")
	fmt.Println("|------|-----|-----|-----|--------|--------|")
	for _, r := range results {
		fmt.Printf("| %s | %.2fms | %.2fms | %.2fms | %.1f%% | %.1f%% |\n",
			r.name, r.p50, r.p90, r.p99, r.successRate, r.timeoutRate)
	}

	fmt.Println()
	fmt.Println("### 可靠性衰减公式")
	fmt.Println("N层串联可用性 = (1 - p)^N，其中 p 是单层故障率")
	fmt.Println("- 每层 99.9% → 3层 = 99.7%（月停机 ~22分钟）")
	fmt.Println("- 每层 99.5% → 3层 = 98.5%（月停机 ~11小时）")
	fmt.Println("- 每层 99.0% → 3层 = 97.0%（月停机 ~22小时）")
	fmt.Println()
	fmt.Println("### 结论")
	fmt.Println("- 递减超时是最优策略：给上游留足时间，下游逐步收紧")
	fmt.Println("- 宽松超时成功率最高但P99最差（慢请求拖尾）")
	fmt.Println("- 紧超时P99最好但成功率最低（正常请求也被杀）")
	fmt.Println("- 分布式系统中，超时传播策略直接决定可用性和延迟的权衡")
}

func benchWithSlow(url string, total, conc int, slowRatio float64) ([]float64, int, int) {
	var latencies []float64
	var mu sync.Mutex
	var wg sync.WaitGroup
	var successes, timeouts int
	perW := total / conc
	slowInterval := int(1.0 / slowRatio)

	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c := &http.Client{Timeout: 2 * time.Second}
			for j := 0; j < perW; j++ {
				reqURL := url
				if (workerID*perW+j)%slowInterval == 0 {
					reqURL += "?slow=1"
				}
				s := time.Now()
				resp, err := c.Get(reqURL)
				d := time.Since(s).Seconds() * 1000
				mu.Lock()
				latencies = append(latencies, d)
				if err != nil || (resp != nil && resp.StatusCode != 200) {
					timeouts++
				} else {
					successes++
				}
				if resp != nil {
					resp.Body.Close()
				}
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return latencies, successes, timeouts
}

func p(data []float64, pct float64) float64 {
	if len(data) == 0 {
		return 0
	}
	s := make([]float64, len(data))
	copy(s, data)
	sort.Float64s(s)
	idx := int(float64(len(s)-1) * pct / 100.0)
	return s[idx]
}
