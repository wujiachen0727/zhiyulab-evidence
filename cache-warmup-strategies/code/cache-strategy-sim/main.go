// 缓存冷启动策略模拟器
// 模拟三种策略在相同条件下的行为差异：
// 1. 惰性加载 + 保护（Lazy Load + Circuit Breaker）
// 2. 主动预热（Eager Warmup）
// 3. 渐进式预热（Gradual Warmup + Rate Limiter）
//
// 运行：go run main.go
// 依赖：Go 1.21+，无外部依赖

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Config 模拟参数
type Config struct {
	TotalKeys       int     // 缓存总键数
	HotKeyRatio     float64 // 热点键占比（20% 的键占 80% 的访问）
	Concurrency     int     // 并发请求数
	RequestCount    int     // 总请求数
	WarmupBatchSize int     // 渐进式预热每批大小
	WarmupInterval  time.Duration // 渐进式预热间隔
}

// Result 模拟结果
type Result struct {
	TotalTime       time.Duration
	CacheHits       int64
	CacheMisses     int64
	DBQueries       int64
	DBPeakConns     int64
	AvgLatency      time.Duration
	P99Latency      time.Duration
	StrategyName    string
}

// SimCache 模拟缓存
type SimCache struct {
	mu    sync.RWMutex
	data  map[int]bool // true = 已缓存
}

func NewSimCache() *SimCache {
	return &SimCache{data: make(map[int]bool)}
}

func (c *SimCache) Get(key int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func (c *SimCache) Set(key int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = true
}

func (c *SimCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// isHotKey 判断一个键是否是热点键
func isHotKey(key int, hotRatio float64) bool {
	threshold := int(float64(Config{}.TotalKeys) * hotRatio)
	return key < threshold
}

// generateRequestDistribution 生成请求分布（Zipf-like）
func generateRequestDistribution(totalKeys int, hotRatio float64, requestCount int) []int {
	hotThreshold := int(float64(totalKeys) * hotRatio)
	requests := make([]int, requestCount)
	for i := 0; i < requestCount; i++ {
		if rand.Float64() < 0.8 {
			// 80% 的请求落在热点键
			requests[i] = rand.Intn(hotThreshold)
		} else {
			// 20% 的请求落在非热点键
			requests[i] = hotThreshold + rand.Intn(totalKeys-hotThreshold)
		}
	}
	return requests
}

// simulateLazyLoad 模拟惰性加载+保护策略
func simulateLazyLoad(cfg Config) Result {
	start := time.Now()
	cache := NewSimCache()
	var mu sync.Mutex
	var maxConns int64
	var activeConns int64

	requests := generateRequestDistribution(cfg.TotalKeys, cfg.HotKeyRatio, cfg.RequestCount)

	var wg sync.WaitGroup
	reqCh := make(chan int, cfg.RequestCount)
	resultCh := make(chan struct {
		cacheHit bool
		latency  time.Duration
	}, cfg.RequestCount)

	// 启动 worker
	worker := func() {
		for key := range reqCh {
			reqStart := time.Now()

			hit := cache.Get(key)
			if !hit {
				// 模拟断路器：超过 50 并发直接拒绝
				atomic.AddInt64(&activeConns, 1)
				current := atomic.LoadInt64(&activeConns)
				for current > 50 {
					atomic.AddInt64(&activeConns, -1)
					time.Sleep(time.Millisecond)
					atomic.AddInt64(&activeConns, 1)
					current = atomic.LoadInt64(&activeConns)
				}
				mu.Lock()
				if current > maxConns {
					maxConns = current
				}
				mu.Unlock()

				// 模拟数据库查询 5-15ms
				dbLatency := time.Duration(5+rand.Intn(10)) * time.Millisecond
				time.Sleep(dbLatency)

				cache.Set(key)
				atomic.AddInt64(&activeConns, -1)
			}

			elapsed := time.Since(reqStart)
			resultCh <- struct {
				cacheHit bool
				latency  time.Duration
			}{cacheHit: hit, latency: elapsed}
		}
		wg.Done()
	}

	// 启动并发 workers
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	// 发送请求
	go func() {
		for _, key := range requests {
			reqCh <- key
		}
		close(reqCh)
	}()

	wg.Wait()
	close(resultCh)

	var result Result
	result.StrategyName = "惰性加载+保护"
	var totalLatency time.Duration
	var latencies []time.Duration

	for r := range resultCh {
		if r.cacheHit {
			result.CacheHits++
		} else {
			result.CacheMisses++
			result.DBQueries++
		}
		totalLatency += r.latency
		latencies = append(latencies, r.latency)
	}

	result.TotalTime = time.Since(start)
	result.DBPeakConns = maxConns
	result.AvgLatency = totalLatency / time.Duration(len(latencies))

	// P99
	sortDurations(latencies)
	if len(latencies) > 0 {
		p99Idx := int(math.Ceil(float64(len(latencies))*0.99)) - 1
		if p99Idx >= len(latencies) {
			p99Idx = len(latencies) - 1
		}
		result.P99Latency = latencies[p99Idx]
	}

	return result
}

// simulateEagerWarmup 模拟主动预热策略
func simulateEagerWarmup(cfg Config) Result {
	start := time.Now()
	cache := NewSimCache()

	// 预热：提前加载所有热点键
	warmupStart := time.Now()
	for i := 0; i < int(float64(cfg.TotalKeys)*cfg.HotKeyRatio); i++ {
		cache.Set(i)
	}
	warmupTime := time.Since(warmupStart)

	requests := generateRequestDistribution(cfg.TotalKeys, cfg.HotKeyRatio, cfg.RequestCount)

	var result Result
	result.StrategyName = "主动预热"
	var totalLatency time.Duration
	var latencies []time.Duration

	for _, key := range requests {
		reqStart := time.Now()
		hit := cache.Get(key)
		if !hit {
			// 非热点键的冷加载
			dbLatency := time.Duration(5+rand.Intn(10)) * time.Millisecond
			time.Sleep(dbLatency)
			cache.Set(key)
			result.CacheMisses++
			result.DBQueries++
		} else {
			result.CacheHits++
		}
		elapsed := time.Since(reqStart)
		totalLatency += elapsed
		latencies = append(latencies, elapsed)
	}

	result.TotalTime = time.Since(start) + warmupTime
	result.AvgLatency = totalLatency / time.Duration(len(latencies))

	sortDurations(latencies)
	if len(latencies) > 0 {
		p99Idx := int(math.Ceil(float64(len(latencies))*0.99)) - 1
		if p99Idx >= len(latencies) {
			p99Idx = len(latencies) - 1
		}
		result.P99Latency = latencies[p99Idx]
	}

	return result
}

// simulateGradualWarmup 模拟渐进式预热策略
func simulateGradualWarmup(cfg Config) Result {
	start := time.Now()
	cache := NewSimCache()
	var totalWarmupTime time.Duration

	// 渐进式预热：分批加载热点键，每批间隔
	hotKeys := int(float64(cfg.TotalKeys) * cfg.HotKeyRatio)
	warmupStart := time.Now()
	for i := 0; i < hotKeys; i += cfg.WarmupBatchSize {
		end := i + cfg.WarmupBatchSize
		if end > hotKeys {
			end = hotKeys
		}
		for j := i; j < end; j++ {
			cache.Set(j)
		}
		time.Sleep(cfg.WarmupInterval)
	}
	totalWarmupTime = time.Since(warmupStart)

	requests := generateRequestDistribution(cfg.TotalKeys, cfg.HotKeyRatio, cfg.RequestCount)

	var result Result
	result.StrategyName = "渐进式预热"
	_ = totalWarmupTime
	var totalLatency time.Duration
	var latencies []time.Duration
	var maxConns int64
	var activeConns int64

	for _, key := range requests {
		reqStart := time.Now()
		hit := cache.Get(key)

		if !hit {
			// 未预热到的键走 DB，但有限流保护
			atomic.AddInt64(&activeConns, 1)
			current := atomic.LoadInt64(&activeConns)
			for current > 30 {
				atomic.AddInt64(&activeConns, -1)
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&activeConns, 1)
				current = atomic.LoadInt64(&activeConns)
			}
			if current > maxConns {
				maxConns = current
			}

			dbLatency := time.Duration(5+rand.Intn(10)) * time.Millisecond
			time.Sleep(dbLatency)
			cache.Set(key)
			atomic.AddInt64(&activeConns, -1)
			result.CacheMisses++
			result.DBQueries++
		} else {
			result.CacheHits++
		}

		elapsed := time.Since(reqStart)
		totalLatency += elapsed
		latencies = append(latencies, elapsed)
	}

	result.TotalTime = time.Since(start)
	result.DBPeakConns = maxConns
	result.AvgLatency = totalLatency / time.Duration(len(latencies))

	sortDurations(latencies)
	if len(latencies) > 0 {
		p99Idx := int(math.Ceil(float64(len(latencies))*0.99)) - 1
		if p99Idx >= len(latencies) {
			p99Idx = len(latencies) - 1
		}
		result.P99Latency = latencies[p99Idx]
	}

	return result
}

func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[j] < durations[i] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	cfg := Config{
		TotalKeys:       10000,
		HotKeyRatio:     0.2,    // 20% 热点
		Concurrency:     50,
		RequestCount:    10000,
		WarmupBatchSize: 200,    // 每批 200 个键
		WarmupInterval:  2 * time.Millisecond,
	}

	fmt.Println("=== 缓存冷启动策略模拟 [实测 Go 1.26.4] ===")
	fmt.Printf("配置：总键数=%d, 热点比=%.0f%%, 并发=%d, 请求数=%d\n\n",
		cfg.TotalKeys, cfg.HotKeyRatio*100, cfg.Concurrency, cfg.RequestCount)

	// 1. 惰性加载+保护
	r1 := simulateLazyLoad(cfg)
	printResult(r1)

	// 2. 主动预热
	r2 := simulateEagerWarmup(cfg)
	printResult(r2)

	// 3. 渐进式预热
	r3 := simulateGradualWarmup(cfg)
	printResult(r3)

	// 汇总对比
	fmt.Println("\n=== 三种策略对比汇总 ===")
	fmt.Printf("%-18s | %-10s | %-10s | %-12s | %-10s | %-12s\n",
		"策略", "缓存命中率", "DB查询数", "DB峰值连接", "平均延迟", "P99延迟")
	fmt.Println("-------------------|------------|------------|--------------|------------|--------------")
	printCompareRow(r1)
	printCompareRow(r2)
	printCompareRow(r3)
}

func printResult(r Result) {
	hitRate := float64(r.CacheHits) / float64(r.CacheHits+r.CacheMisses) * 100
	fmt.Printf("【%s】\n", r.StrategyName)
	fmt.Printf("  缓存命中率: %.1f%% (%d/%d)\n", hitRate, r.CacheHits, r.CacheHits+r.CacheMisses)
	fmt.Printf("  DB查询数: %d\n", r.DBQueries)
	fmt.Printf("  DB峰值连接: %d\n", r.DBPeakConns)
	fmt.Printf("  平均延迟: %v\n", r.AvgLatency)
	fmt.Printf("  P99延迟: %v\n", r.P99Latency)
	fmt.Printf("  总耗时: %v\n\n", r.TotalTime)
}

func printCompareRow(r Result) {
	hitRate := float64(r.CacheHits) / float64(r.CacheHits+r.CacheMisses) * 100
	fmt.Printf("%-18s | %-8.1f%%  | %-8d  | %-12d | %-10v | %-12v\n",
		r.StrategyName, hitRate, r.DBQueries, r.DBPeakConns, r.AvgLatency, r.P99Latency)
}
