package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sort"
	"sync"
	"time"
)

// 模拟单节点延迟分布：1% 概率超时（>1s），其余 10-100ms
func handler(w http.ResponseWriter, r *http.Request) {
	latency := 0
	if rand.Float64() < 0.01 {
		latency = 1000 + rand.Intn(1000)
	} else {
		latency = 10 + rand.Intn(90)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Fprintf(w, "ok\n")
}

func fanoutCall(n int, client *http.Client) (time.Duration, error) {
	start := time.Now()
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := client.Get("http://127.0.0.1:8300/work")
			if err != nil {
				errs[idx] = err
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	for _, e := range errs {
		if e != nil {
			return elapsed, e
		}
	}
	return elapsed, nil
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

func runFanoutScenario(fanout, samples int) []time.Duration {
	client := &http.Client{Timeout: 10 * time.Second}
	latencies := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		elapsed, err := fanoutCall(fanout, client)
		if err != nil {
			continue
		}
		latencies = append(latencies, elapsed)
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	return latencies
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", handler)
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	server := &http.Server{Addr: ":8300", Handler: mux}
	go func() {
		log.Println("server on :8300")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 50; i++ {
		resp, _ := http.Get("http://127.0.0.1:8300/work")
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	time.Sleep(1 * time.Second)

	fanouts := []int{1, 10, 50, 100}
	const samples = 200

	fmt.Println("=== Fan-out 放大效应实测 ===")
	fmt.Printf("单节点延迟分布：1%% 长尾（1-2s），99%% 正常（10-100ms）\n")
	fmt.Printf("采样数：%d/场景\n\n", samples)

	fmt.Printf("%-8s %12s %12s %12s %12s\n", "Fan-out", "P50", "P95", "P99", "理论超时%")
	for _, n := range fanouts {
		latencies := runFanoutScenario(n, samples)
		// 理论：P(至少一个节点长尾) = 1 - 0.99^n
		theoryPercent := (1.0 - powSlow(0.99, float64(n))) * 100
		fmt.Printf("%-8d %12v %12v %12v %11.2f%%\n",
			n,
			percentile(latencies, 0.50),
			percentile(latencies, 0.95),
			percentile(latencies, 0.99),
			theoryPercent)
	}

	fmt.Println("\n=== 解读 ===")
	fmt.Println("Fan-out=1: P99 接近单节点长尾（~1-2s）")
	fmt.Println("Fan-out=10: P99 开始放大（任一节点慢即整体慢）")
	fmt.Println("Fan-out=100: P99 接近 100%（理论 63.4%），几乎每次都有节点长尾")

	server.Shutdown(context.Background())
}

func powSlow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
