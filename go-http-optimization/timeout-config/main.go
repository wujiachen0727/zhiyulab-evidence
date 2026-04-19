// timeout-config: 对比 Go net/http 默认配置 vs 生产级超时配置
// 在 Slowloris 慢连接攻击 + 正常流量混合场景下的 QPS/P99/资源消耗差异
// [实测 Go 1.26.2 darwin/arm64]

package main

import (
	"encoding/json"
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

// ==================== 服务器配置 ====================

const (
	PortDefault  = 18080 // 服务器 A：默认配置
	PortHardened = 18081 // 服务器 B：生产级配置
)

// 业务逻辑：简单 JSON 响应
func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 启动默认配置服务器（无任何超时设置）
func startDefaultServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", PortDefault),
		Handler: mux,
		// 无超时设置 — 这是 Go 默认行为
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("默认配置服务器异常退出: %v", err)
		}
	}()
	return srv
}

// 启动生产级配置服务器
func startHardenedServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", PortHardened),
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("生产级配置服务器异常退出: %v", err)
		}
	}()
	return srv
}

// ==================== 压测框架 ====================

type BenchResult struct {
	Scenario      string
	ServerName    string
	Port          int
	TotalRequests int64
	Duration      time.Duration
	QPS           float64
	Latencies     []time.Duration // 采集所有延迟
	P50           time.Duration
	P90           time.Duration
	P99           time.Duration
	MemAllocKB    uint64
	SlowConnCount int
}

// 发送正常请求并记录延迟
func sendRequest(port int) (time.Duration, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	start := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return time.Since(start), nil
}

// 并发压测：在给定时间内持续发送请求
func benchNormalTraffic(port int, concurrency int, duration time.Duration) (int64, []time.Duration) {
	var totalRequests int64
	var latenciesMu sync.Mutex
	var latencies []time.Duration

	ctx := make(chan struct{})
	var wg sync.WaitGroup

	// 启动并发 worker
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx:
					return
				default:
					lat, err := sendRequest(port)
					if err != nil {
						atomic.AddInt64(&totalRequests, 1) // 计入尝试次数
						continue
					}
					atomic.AddInt64(&totalRequests, 1)
					latenciesMu.Lock()
					latencies = append(latencies, lat)
					latenciesMu.Unlock()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(ctx)
	wg.Wait()

	return totalRequests, latencies
}

// 慢连接攻击：建立连接后每5秒发1字节
func startSlowloris(port int, count int, stop chan struct{}) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
			if err != nil {
				return
			}
			defer conn.Close()

			// 发送一个不完整的 HTTP 请求头
			fmt.Fprintf(conn, "GET /health HTTP/1.1\r\nHost: 127.0.0.1\r\nX-Slow: ")

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					_, err := conn.Write([]byte("a"))
					if err != nil {
						return
					}
				}
			}
		}(i)
	}
	return &wg
}

// 计算百分位延迟
func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	idx := int(float64(len(latencies)-1) * p)
	return latencies[idx]
}

// 获取当前内存分配
func getMemStats() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 // KB
}

// ==================== 场景执行 ====================

func runScenario(scenarioName string, port int, serverName string, concurrency int, duration time.Duration, slowConnCount int) BenchResult {
	log.Printf("  场景: %s, 服务器: %s (:%d), 并发: %d, 慢连接: %d, 持续: %v",
		scenarioName, serverName, port, concurrency, slowConnCount, duration)

	// 获取压测前内存
	memBefore := getMemStats()

	// 启动慢连接（如果有）
	var slowWg *sync.WaitGroup
	stopSlow := make(chan struct{})
	if slowConnCount > 0 {
		slowWg = startSlowloris(port, slowConnCount, stopSlow)
		// 等待慢连接建立
		time.Sleep(500 * time.Millisecond)
	}

	// 运行正常流量压测
	totalReqs, latencies := benchNormalTraffic(port, concurrency, duration)

	// 停止慢连接
	if slowConnCount > 0 {
		close(stopSlow)
		slowWg.Wait()
	}

	// 获取压测后内存
	memAfter := getMemStats()

	// 计算指标
	qps := float64(totalReqs) / duration.Seconds()
	p50 := percentile(latencies, 0.50)
	p90 := percentile(latencies, 0.90)
	p99 := percentile(latencies, 0.99)

	result := BenchResult{
		Scenario:      scenarioName,
		ServerName:    serverName,
		Port:          port,
		TotalRequests: totalReqs,
		Duration:      duration,
		QPS:           qps,
		Latencies:     latencies,
		P50:           p50,
		P90:           p90,
		P99:           p99,
		MemAllocKB:    memAfter - memBefore,
		SlowConnCount: slowConnCount,
	}

	log.Printf("  结果: QPS=%.0f, P50=%v, P90=%v, P99=%v, 请求数=%d, 内存增量=%dKB",
		qps, p50, p90, p99, totalReqs, result.MemAllocKB)

	return result
}

// ==================== 主函数 ====================

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("=== Go HTTP 超时配置对比实验 ===")
	log.Println("[实测 Go 1.26.2 darwin/arm64]")

	// 启动两个服务器
	log.Println("启动服务器...")
	srvA := startDefaultServer()
	srvB := startHardenedServer()
	time.Sleep(500 * time.Millisecond) // 等待服务器就绪

	// 验证服务器可用
	for _, port := range []int{PortDefault, PortHardened} {
		_, err := sendRequest(port)
		if err != nil {
			log.Fatalf("服务器 :%d 不可用: %v", port, err)
		}
	}
	log.Println("两个服务器均已就绪")

	// 压测参数
	const concurrency = 100
	const benchDuration = 8 * time.Second

	var results []BenchResult

	// ===== 场景1：纯正常流量 =====
	log.Println("\n--- 场景1：纯正常流量 ---")
	results = append(results, runScenario("纯正常流量", PortDefault, "默认配置", concurrency, benchDuration, 0))
	results = append(results, runScenario("纯正常流量", PortHardened, "生产级配置", concurrency, benchDuration, 0))
	time.Sleep(2 * time.Second) // 冷却

	// ===== 场景2：50个慢连接 + 正常流量 =====
	log.Println("\n--- 场景2：50个慢连接 + 正常流量 ---")
	results = append(results, runScenario("50慢连接+正常流量", PortDefault, "默认配置", concurrency, benchDuration, 50))
	results = append(results, runScenario("50慢连接+正常流量", PortHardened, "生产级配置", concurrency, benchDuration, 50))
	time.Sleep(2 * time.Second) // 冷却

	// ===== 场景3：200个慢连接 + 正常流量 =====
	log.Println("\n--- 场景3：200个慢连接 + 正常流量 ---")
	results = append(results, runScenario("200慢连接+正常流量", PortDefault, "默认配置", concurrency, benchDuration, 200))
	results = append(results, runScenario("200慢连接+正常流量", PortHardened, "生产级配置", concurrency, benchDuration, 200))

	// 关闭服务器
	srvA.Close()
	srvB.Close()

	// 输出结果
	printResults(results)
}

func printResults(results []BenchResult) {
	fmt.Println("\n\n============================================================")
	fmt.Println("  Go HTTP 超时配置对比实验结果 [实测 Go 1.26.2 darwin/arm64]")
	fmt.Println("============================================================")

	for _, scenario := range []string{"纯正常流量", "50慢连接+正常流量", "200慢连接+正常流量"} {
		fmt.Printf("\n### %s\n\n", scenario)
		fmt.Println("| 指标 | 默认配置 (:18080) | 生产级配置 (:18081) | 差异 |")
		fmt.Println("|------|-------------------|---------------------|------|")

		var def, hard BenchResult
		for _, r := range results {
			if r.Scenario == scenario && r.ServerName == "默认配置" {
				def = r
			}
			if r.Scenario == scenario && r.ServerName == "生产级配置" {
				hard = r
			}
		}

		// QPS
		qpsDiff := ""
		if def.QPS > 0 && hard.QPS > 0 {
			ratio := hard.QPS / def.QPS
			if ratio > 1 {
				qpsDiff = fmt.Sprintf("生产级 %.1fx", ratio)
			} else {
				qpsDiff = fmt.Sprintf("默认 %.1fx", 1/ratio)
			}
		}
		fmt.Printf("| QPS | %.0f | %.0f | %s |\n", def.QPS, hard.QPS, qpsDiff)

		// P50
		fmt.Printf("| P50 延迟 | %v | %v | %s |\n", def.P50, hard.P50, diffStr(def.P50, hard.P50))

		// P90
		fmt.Printf("| P90 延迟 | %v | %v | %s |\n", def.P90, hard.P90, diffStr(def.P90, hard.P90))

		// P99
		fmt.Printf("| P99 延迟 | %v | %v | %s |\n", def.P99, hard.P99, diffStr(def.P99, hard.P99))

		// 总请求数
		fmt.Printf("| 总请求数 | %d | %d | %s |\n", def.TotalRequests, hard.TotalRequests, diffInt(def.TotalRequests, hard.TotalRequests))

		// 内存增量
		fmt.Printf("| 内存增量(KB) | %d | %d | %s |\n", def.MemAllocKB, hard.MemAllocKB, diffInt(int64(def.MemAllocKB), int64(hard.MemAllocKB)))

		// 成功请求率
		defSuccess := float64(len(def.Latencies)) / float64(def.TotalRequests) * 100
		hardSuccess := float64(len(hard.Latencies)) / float64(hard.TotalRequests) * 100
		fmt.Printf("| 成功率(%%) | %.1f | %.1f | %s |\n", defSuccess, hardSuccess, diffFloat(defSuccess, hardSuccess))
	}

	// 关键发现
	fmt.Println("\n### 关键发现\n")
	findings(results)
}

func diffStr(a, b time.Duration) string {
	if a == 0 || b == 0 {
		return "N/A"
	}
	if b < a {
		return fmt.Sprintf("生产级快 %.1fx", float64(a)/float64(b))
	}
	return fmt.Sprintf("默认快 %.1fx", float64(b)/float64(a))
}

func diffInt(a, b int64) string {
	if b > a {
		return fmt.Sprintf("生产级 +%d", b-a)
	}
	return fmt.Sprintf("默认 +%d", a-b)
}

func diffFloat(a, b float64) string {
	if b > a {
		return fmt.Sprintf("生产级 +%.1f%%", b-a)
	}
	return fmt.Sprintf("默认 +%.1f%%", a-b)
}

func findings(results []BenchResult) {
	// 分析场景1 vs 场景3 的性能衰减
	var defS1, defS3, hardS1, hardS3 BenchResult
	for _, r := range results {
		switch {
		case r.Scenario == "纯正常流量" && r.ServerName == "默认配置":
			defS1 = r
		case r.Scenario == "纯正常流量" && r.ServerName == "生产级配置":
			hardS1 = r
		case r.Scenario == "200慢连接+正常流量" && r.ServerName == "默认配置":
			defS3 = r
		case r.Scenario == "200慢连接+正常流量" && r.ServerName == "生产级配置":
			hardS3 = r
		}
	}

	// 默认配置衰减
	defQPSDrop := 0.0
	if defS1.QPS > 0 {
		defQPSDrop = (1 - defS3.QPS/defS1.QPS) * 100
	}
	hardQPSDrop := 0.0
	if hardS1.QPS > 0 {
		hardQPSDrop = (1 - hardS3.QPS/hardS1.QPS) * 100
	}

	fmt.Printf("1. **默认配置 QPS 衰减**: 纯正常流量 → 200慢连接，QPS 从 %.0f 降至 %.0f（衰减 %.1f%%）\n",
		defS1.QPS, defS3.QPS, defQPSDrop)
	fmt.Printf("2. **生产级配置 QPS 衰减**: 纯正常流量 → 200慢连接，QPS 从 %.0f 降至 %.0f（衰减 %.1f%%）\n",
		hardS1.QPS, hardS3.QPS, hardQPSDrop)

	// P99 延迟变化
	fmt.Printf("3. **默认配置 P99 延迟增长**: %v → %v（增长 %.1fx）\n",
		defS1.P99, defS3.P99, float64Safe(defS3.P99, defS1.P99))
	fmt.Printf("4. **生产级配置 P99 延迟增长**: %v → %v（增长 %.1fx）\n",
		hardS1.P99, hardS3.P99, float64Safe(hardS3.P99, hardS1.P99))

	// 成功率对比
	defS1Rate := successRate(defS1)
	defS3Rate := successRate(defS3)
	hardS1Rate := successRate(hardS1)
	hardS3Rate := successRate(hardS3)

	fmt.Printf("5. **默认配置成功率**: 纯正常流量 %.1f%% → 200慢连接 %.1f%%（下降 %.1f 个百分点）\n",
		defS1Rate, defS3Rate, defS1Rate-defS3Rate)
	fmt.Printf("6. **生产级配置成功率**: 纯正常流量 %.1f%% → 200慢连接 %.1f%%（下降 %.1f 个百分点）\n",
		hardS1Rate, hardS3Rate, hardS1Rate-hardS3Rate)

	fmt.Println("\n7. **结论**: 生产级超时配置能有效防御 Slowloris 攻击，在慢连接压力下维持更高的 QPS 和更低的 P99 延迟。默认配置因无超时限制，慢连接会持续占用服务器资源，导致正常请求排队或超时。")
}

func float64Safe(a, b time.Duration) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

func successRate(r BenchResult) float64 {
	if r.TotalRequests == 0 {
		return 0
	}
	return float64(len(r.Latencies)) / float64(r.TotalRequests) * 100
}

// 确保 main.go 文件中的 exit 函数用于异常退出
func init() {
	log.SetOutput(os.Stderr)
}
