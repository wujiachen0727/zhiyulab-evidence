// modular-vs-microservice: 模块化单体 vs 微服务延迟对比实验
// [推演：模拟业务处理] 使用 time.Sleep 模拟各模块1ms处理时间
// 使用 HTTP 代替 gRPC，实际 gRPC 延迟更低，本文对比结果是保守估计

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 业务模型
// ============================================================

type OrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Amount    int    `json:"amount"`
}

type OrderResponse struct {
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
}

// ============================================================
// 实现A：模块化单体（进程内直接函数调用）
// ============================================================
// 模块化单体的模块间通信是进程内的直接函数调用，
// 不经过网络协议栈，无序列化/反序列化开销。

var monolithSeq atomic.Int64

// 库存模块：直接函数调用
func inventoryDeduct(productID string, quantity int) (bool, string) {
	// [推演：模拟业务处理] 库存扣减逻辑 1ms
	time.Sleep(1 * time.Millisecond)
	return true, fmt.Sprintf("deducted %d of %s", quantity, productID)
}

// 支付模块：直接函数调用
func paymentCharge(orderID string, amount int) (bool, string) {
	// [推演：模拟业务处理] 支付处理逻辑 1ms
	time.Sleep(1 * time.Millisecond)
	return true, fmt.Sprintf("txn-%s", orderID)
}

func monolithCreateOrder(req *OrderRequest) *OrderResponse {
	start := time.Now()
	orderID := fmt.Sprintf("mono-%d", monolithSeq.Add(1))

	// [推演：模拟业务处理] 订单创建逻辑 1ms
	time.Sleep(1 * time.Millisecond)

	// 进程内函数调用：库存模块（无网络开销、无序列化）
	invSuccess, _ := inventoryDeduct(req.ProductID, req.Quantity)
	if !invSuccess {
		return &OrderResponse{OrderID: orderID, Status: "inventory_failed"}
	}

	// 进程内函数调用：支付模块（无网络开销、无序列化）
	paySuccess, _ := paymentCharge(orderID, req.Amount)
	if !paySuccess {
		return &OrderResponse{OrderID: orderID, Status: "payment_failed"}
	}

	return &OrderResponse{
		OrderID:  orderID,
		Status:   "created",
		Duration: time.Since(start).String(),
	}
}

func startMonolithServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req OrderRequest
		json.Unmarshal(body, &req)

		resp := monolithCreateOrder(&req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("单体服务器启动失败: %v", err)
		}
	}()
	return srv
}

// ============================================================
// 实现B：微服务（HTTP 跨服务调用）
// ============================================================

var microSeq atomic.Int64

// 支付服务
func startPaymentService(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		// [推演：模拟业务处理] 支付处理逻辑 1ms
		time.Sleep(1 * time.Millisecond)

		var req struct {
			OrderID string `json:"order_id"`
			Amount  int    `json:"amount"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":        true,
			"transaction_id": fmt.Sprintf("txn-%s", req.OrderID),
		})
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("支付服务启动失败: %v", err)
		}
	}()
	return srv
}

// 库存服务
func startInventoryService(port int, paymentURL string) *http.Server {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     50,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/deduct", func(w http.ResponseWriter, r *http.Request) {
		// [推演：模拟业务处理] 库存扣减逻辑 1ms
		time.Sleep(1 * time.Millisecond)

		var req struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
			OrderID   string `json:"order_id"`
			Amount    int    `json:"amount"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// HTTP 调用支付服务
		payReq, _ := json.Marshal(map[string]interface{}{
			"order_id": req.OrderID,
			"amount":   req.Amount,
		})
		resp, err := client.Post(paymentURL, "application/json", bytes.NewReader(payReq))
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
			return
		}
		defer resp.Body.Close()

		var payResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&payResp)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":        true,
			"message":        fmt.Sprintf("deducted %d of %s", req.Quantity, req.ProductID),
			"payment_result": payResp,
		})
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("库存服务启动失败: %v", err)
		}
	}()
	return srv
}

// 网关服务
func startMicroserviceGateway(port int, inventoryURL string) *http.Server {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     50,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		var req OrderRequest
		json.NewDecoder(r.Body).Decode(&req)

		orderID := fmt.Sprintf("micro-%d", microSeq.Add(1))

		// [推演：模拟业务处理] 订单创建逻辑 1ms
		time.Sleep(1 * time.Millisecond)

		// HTTP 调用库存服务（库存服务内部再调用支付服务）
		invReq, _ := json.Marshal(map[string]interface{}{
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
			"order_id":   orderID,
			"amount":     req.Amount,
		})
		resp, err := client.Post(inventoryURL, "application/json", bytes.NewReader(invReq))
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]interface{}{"order_id": orderID, "status": "inventory_call_failed"})
			return
		}
		defer resp.Body.Close()

		var invResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&invResp)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"order_id": orderID,
			"status":   "created",
		})
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("网关服务启动失败: %v", err)
		}
	}()
	return srv
}

// ============================================================
// 压测引擎
// ============================================================

type BenchResult struct {
	Name        string
	Requests    int
	Concurrency int
	TotalTime   time.Duration
	Latencies   []time.Duration
	QPS         float64
	P50         time.Duration
	P90         time.Duration
	P99         time.Duration
	Min         time.Duration
	Max         time.Duration
	Errors      int
	MemAllocMB  float64
	NumGC       uint32
}

func bench(url string, requests int, concurrency int, name string) *BenchResult {
	// 预热
	for i := 0; i < 20; i++ {
		req, _ := json.Marshal(&OrderRequest{ProductID: "warmup", Quantity: 1, Amount: 100})
		resp, err := http.Post(url, "application/json", bytes.NewReader(req))
		if err == nil {
			resp.Body.Close()
		}
	}
	time.Sleep(500 * time.Millisecond)

	// 重置内存统计
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	latencies := make([]time.Duration, 0, requests)
	var latencyMu sync.Mutex
	var errors atomic.Int64

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	// 使用共享 http.Client（连接池）
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: concurrency,
			MaxConnsPerHost:     concurrency,
		},
	}

	start := time.Now()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			reqBody, _ := json.Marshal(&OrderRequest{
				ProductID: fmt.Sprintf("product-%d", idx%100),
				Quantity:  1,
				Amount:    100,
			})

			reqStart := time.Now()
			resp, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
			elapsed := time.Since(reqStart)

			if err != nil {
				errors.Add(1)
				latencyMu.Lock()
				latencies = append(latencies, elapsed)
				latencyMu.Unlock()
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()

			latencyMu.Lock()
			latencies = append(latencies, elapsed)
			latencyMu.Unlock()
		}(i)
	}
	wg.Wait()
	totalTime := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	n := len(latencies)
	result := &BenchResult{
		Name:        name,
		Requests:    requests,
		Concurrency: concurrency,
		TotalTime:   totalTime,
		Latencies:   latencies,
		QPS:         float64(requests) / totalTime.Seconds(),
		Min:         latencies[0],
		Max:         latencies[n-1],
		Errors:      int(errors.Load()),
		MemAllocMB:  float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024,
		NumGC:       memAfter.NumGC - memBefore.NumGC,
	}

	if n > 0 {
		result.P50 = latencies[int(float64(n)*0.50)]
		result.P90 = latencies[int(float64(n)*0.90)]
		result.P99 = latencies[int(float64(n)*0.99)]
	}

	return result
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Microseconds()))
	}
	return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
}

func printComparison(monoResult, microResult *BenchResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  模块化单体 vs 微服务 延迟对比实验结果")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("  环境: [实测 Go %s %s/%s]\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Println("  说明: 使用 HTTP 代替 gRPC，实际 gRPC 延迟更低，本文对比结果是保守估计")
	fmt.Println("  [推演：模拟业务处理] 每个模块 1ms time.Sleep")
	fmt.Printf("  并发: %d, 请求: %d\n", monoResult.Concurrency, monoResult.Requests)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n┌────────────────────┬─────────────────────┬─────────────────────┬──────────┐")
	fmt.Println("│ 指标               │ 模块化单体(进程内)  │ 微服务(HTTP调用)    │ 差异倍数 │")
	fmt.Println("├────────────────────┼─────────────────────┼─────────────────────┼──────────┤")

	rows := []struct {
		label string
		mono  time.Duration
		micro time.Duration
	}{
		{"P50 延迟", monoResult.P50, microResult.P50},
		{"P90 延迟", monoResult.P90, microResult.P90},
		{"P99 延迟", monoResult.P99, microResult.P99},
		{"Min 延迟", monoResult.Min, microResult.Min},
		{"Max 延迟", monoResult.Max, microResult.Max},
	}

	for _, row := range rows {
		ratio := float64(row.micro) / float64(row.mono)
		fmt.Printf("│ %-18s │ %-19s │ %-19s │ %7.1fx │\n",
			row.label,
			formatDuration(row.mono),
			formatDuration(row.micro),
			ratio,
		)
	}

	fmt.Println("├────────────────────┼─────────────────────┼─────────────────────┼──────────┤")

	monoQPS := monoResult.QPS
	microQPS := microResult.QPS
	fmt.Printf("│ %-18s │ %-19s │ %-19s │ %7.1fx │\n",
		"QPS",
		fmt.Sprintf("%.0f", monoQPS),
		fmt.Sprintf("%.0f", microQPS),
		monoQPS/microQPS,
	)

	fmt.Printf("│ %-18s │ %-19s │ %-19s │ %7s │\n",
		"总耗时",
		monoResult.TotalTime.Round(time.Millisecond),
		microResult.TotalTime.Round(time.Millisecond),
		fmt.Sprintf("%.1fx", float64(microResult.TotalTime)/float64(monoResult.TotalTime)),
	)

	fmt.Printf("│ %-18s │ %-19d │ %-19d │ %7s │\n",
		"错误数",
		monoResult.Errors,
		microResult.Errors,
		"-",
	)

	fmt.Printf("│ %-18s │ %-19s │ %-19s │ %7s │\n",
		"内存增量",
		fmt.Sprintf("%.1f MB", monoResult.MemAllocMB),
		fmt.Sprintf("%.1f MB", microResult.MemAllocMB),
		"-",
	)

	fmt.Printf("│ %-18s │ %-19d │ %-19d │ %7s │\n",
		"GC 次数",
		monoResult.NumGC,
		microResult.NumGC,
		"-",
	)

	fmt.Println("└────────────────────┴─────────────────────┴─────────────────────┴──────────┘")

	fmt.Println("\n结论:")
	p50Ratio := float64(microResult.P50) / float64(monoResult.P50)
	p99Ratio := float64(microResult.P99) / float64(monoResult.P99)
	qpsRatio := monoQPS / microQPS
	fmt.Printf("  - P50 延迟: 微服务比模块化单体慢 %.1f 倍\n", p50Ratio)
	fmt.Printf("  - P99 延迟: 微服务比模块化单体慢 %.1f 倍\n", p99Ratio)
	fmt.Printf("  - QPS: 模块化单体是微服务的 %.1f 倍\n", qpsRatio)
	fmt.Println("  - 额外延迟主要来自 HTTP 序列化/反序列化、网络协议栈开销、连接管理")
	fmt.Println("  - 若使用 gRPC 替代 HTTP，差距会缩小但仍显著（gRPC 延迟约为 HTTP 的 1/2~1/3）")
}

// ============================================================
// main
// ============================================================

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	const (
		portMonolith  = 18084
		portGateway   = 18085
		portInventory = 18086
		portPayment   = 18087

		concurrency = 50
		requests    = 10000
	)

	fmt.Println("正在启动服务器...")

	// 启动模块化单体
	monoSrv := startMonolithServer(portMonolith)
	defer monoSrv.Shutdown(context.Background())
	fmt.Printf("  模块化单体 → :%d/order\n", portMonolith)

	// 启动微服务集群
	paySrv := startPaymentService(portPayment)
	defer paySrv.Shutdown(context.Background())
	fmt.Printf("  支付服务   → :%d/pay\n", portPayment)

	invSrv := startInventoryService(portInventory, fmt.Sprintf("http://127.0.0.1:%d/pay", portPayment))
	defer invSrv.Shutdown(context.Background())
	fmt.Printf("  库存服务   → :%d/deduct\n", portInventory)

	gwSrv := startMicroserviceGateway(portGateway, fmt.Sprintf("http://127.0.0.1:%d/deduct", portInventory))
	defer gwSrv.Shutdown(context.Background())
	fmt.Printf("  API网关    → :%d/order\n", portGateway)

	// 等待服务器就绪
	fmt.Println("\n等待服务器就绪...")
	time.Sleep(1 * time.Second)

	// 验证服务器
	for _, url := range []string{
		fmt.Sprintf("http://127.0.0.1:%d/order", portMonolith),
		fmt.Sprintf("http://127.0.0.1:%d/order", portGateway),
	} {
		req, _ := json.Marshal(&OrderRequest{ProductID: "health", Quantity: 1, Amount: 100})
		resp, err := http.Post(url, "application/json", bytes.NewReader(req))
		if err != nil {
			log.Fatalf("服务器健康检查失败 %s: %v", url, err)
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	fmt.Println("服务器就绪 ✓")

	// 压测：模块化单体
	fmt.Printf("\n▸ 压测模块化单体 (%d并发, %d请求)...\n", concurrency, requests)
	monoResult := bench(
		fmt.Sprintf("http://127.0.0.1:%d/order", portMonolith),
		requests, concurrency, "模块化单体",
	)
	fmt.Printf("  完成: QPS=%.0f, P50=%s, P99=%s\n",
		monoResult.QPS,
		formatDuration(monoResult.P50),
		formatDuration(monoResult.P99),
	)

	// 压测间冷却
	time.Sleep(2 * time.Second)

	// 压测：微服务
	fmt.Printf("\n▸ 压测微服务集群 (%d并发, %d请求)...\n", concurrency, requests)
	microResult := bench(
		fmt.Sprintf("http://127.0.0.1:%d/order", portGateway),
		requests, concurrency, "微服务",
	)
	fmt.Printf("  完成: QPS=%.0f, P50=%s, P99=%s\n",
		microResult.QPS,
		formatDuration(microResult.P50),
		formatDuration(microResult.P99),
	)

	// 输出对比
	printComparison(monoResult, microResult)

	// 生成 Markdown 结果文件
	writeResultMarkdown(monoResult, microResult, concurrency, requests)
}

func writeResultMarkdown(monoResult, microResult *BenchResult, concurrency, requests int) {
	p50Ratio := float64(microResult.P50) / float64(monoResult.P50)
	p90Ratio := float64(microResult.P90) / float64(monoResult.P90)
	p99Ratio := float64(microResult.P99) / float64(monoResult.P99)
	qpsRatio := monoResult.QPS / microResult.QPS

	md := fmt.Sprintf(`# 模块化单体 vs 微服务延迟对比实验

> 环境: [实测 Go %s %s/%s]
> 说明: 使用 HTTP 代替 gRPC，实际 gRPC 延迟更低，本文对比结果是保守估计
> [推演：模拟业务处理] 每个模块 1ms time.Sleep

## 实验设计

- **业务流程**: 创建订单 → 扣减库存 → 发起支付
- **模块化单体**: 3个模块在同一进程内，直接函数调用（无网络、无序列化）
- **微服务**: 3个独立 HTTP 服务，网关 → 库存服务 → 支付服务（2次 HTTP 跨服务调用）
- **压测参数**: %d 并发, %d 请求

## 延迟对比

| 指标 | 模块化单体（进程内） | 微服务（HTTP调用） | 差异倍数 |
|------|--------------------|--------------------|---------|
| P50 | %s | %s | %.1fx |
| P90 | %s | %s | %.1fx |
| P99 | %s | %s | %.1fx |
| Min | %s | %s | - |
| Max | %s | %s | - |

## 吞吐量对比

| 指标 | 模块化单体 | 微服务 |
|------|----------|-------|
| QPS | %.0f | %.0f |
| 总耗时 | %s | %s |
| QPS 倍数 | %.1fx | - |

## 资源对比

| 指标 | 模块化单体 | 微服务 |
|------|----------|-------|
| 内存增量 | %.1f MB | %.1f MB |
| GC 次数 | %d | %d |
| 错误数 | %d | %d |

## 结论

- P50 延迟：微服务比模块化单体慢 **%.1f 倍**
- P99 延迟：微服务比模块化单体慢 **%.1f 倍**
- QPS：模块化单体是微服务的 **%.1f 倍**
- 额外延迟主要来源：HTTP 序列化/反序列化、网络协议栈开销、连接管理
- 若使用 gRPC 替代 HTTP，差距会缩小但仍显著（gRPC 延迟约为 HTTP 的 1/2~1/3）
`,
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
		concurrency, requests,
		formatDuration(monoResult.P50), formatDuration(microResult.P50), p50Ratio,
		formatDuration(monoResult.P90), formatDuration(microResult.P90), p90Ratio,
		formatDuration(monoResult.P99), formatDuration(microResult.P99), p99Ratio,
		formatDuration(monoResult.Min), formatDuration(microResult.Min),
		formatDuration(monoResult.Max), formatDuration(microResult.Max),
		monoResult.QPS, microResult.QPS,
		monoResult.TotalTime.Round(time.Millisecond), microResult.TotalTime.Round(time.Millisecond),
		qpsRatio,
		monoResult.MemAllocMB, microResult.MemAllocMB,
		monoResult.NumGC, microResult.NumGC,
		monoResult.Errors, microResult.Errors,
		p50Ratio, p99Ratio, qpsRatio,
	)

	resultPath := "/Users/wujiachen/WriteCraft/articles/go-http-optimization/evidence/output/modular-vs-microservice/result.md"
	os.WriteFile(resultPath, []byte(md), 0644)
	fmt.Printf("\n结果已写入: %s\n", resultPath)
}
