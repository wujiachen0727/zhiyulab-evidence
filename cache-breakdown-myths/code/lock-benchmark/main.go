package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// 模拟从数据库获取数据（固定延迟）
func fetchFromDB(key string) (string, error) {
	time.Sleep(50 * time.Millisecond) // 模拟 50ms DB 查询
	return fmt.Sprintf("value_for_%s", key), nil
}

// 方案 1：互斥锁（本地 sync.Mutex）
func withMutex(key string, mu *sync.Mutex, cache map[string]string) (string, time.Duration) {
	start := time.Now()
	mu.Lock()
	if v, ok := cache[key]; ok {
		mu.Unlock()
		return v, time.Since(start)
	}
	v, _ := fetchFromDB(key)
	cache[key] = v
	mu.Unlock()
	return v, time.Since(start)
}

// 方案 2：singleflight
func withSingleflight(key string, g *singleflight.Group) (string, time.Duration) {
	start := time.Now()
	v, _, _ := g.Do(key, func() (interface{}, error) {
		return fetchFromDB(key)
	})
	return v.(string), time.Since(start)
}

// 方案 3：逻辑过期（模拟：缓存永不过期，后台异步刷新）
type logicalExpireCache struct {
	mu       sync.RWMutex
	data     map[string]string
	expireAt map[string]time.Time
}

func newLogicalExpireCache() *logicalExpireCache {
	return &logicalExpireCache{
		data:     make(map[string]string),
		expireAt: make(map[string]time.Time),
	}
}

func (c *logicalExpireCache) get(key string) (string, time.Duration) {
	start := time.Now()
	c.mu.RLock()
	v, ok := c.data[key]
	exp, _ := c.expireAt[key]
	c.mu.RUnlock()

	if ok {
		if time.Now().After(exp) {
			// 逻辑过期，后台异步刷新（fire and forget）
			go func() {
				newV, _ := fetchFromDB(key)
				c.mu.Lock()
				c.data[key] = newV
				c.expireAt[key] = time.Now().Add(10 * time.Second)
				c.mu.Unlock()
			}()
		}
		return v, time.Since(start) // 返回旧值，不阻塞
	}
	// 冷启动：首次加载
	newV, _ := fetchFromDB(key)
	c.mu.Lock()
	c.data[key] = newV
	c.expireAt[key] = time.Now().Add(10 * time.Second)
	c.mu.Unlock()
	return newV, time.Since(start)
}

func runConcurrentTest(name string, concurrency int, fn func(key string) time.Duration) {
	var wg sync.WaitGroup
	latencies := make([]time.Duration, concurrency)
	var totalBlocked int64

	// 预热：清空缓存状态，让所有请求同时到达
	barrier := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-barrier // 等待统一放行
			d := fn("hot_key")
			latencies[idx] = d
			if d > 55*time.Millisecond {
				// 等了超过 DB 查询时间 = 被阻塞
				atomic.AddInt64(&totalBlocked, 1)
			}
		}(i)
	}

	time.Sleep(10 * time.Millisecond) // 确保所有 goroutine 就绪
	close(barrier)                     // 放行
	wg.Wait()

	// 统计
	var sum time.Duration
	var max time.Duration
	var min = time.Hour
	for _, d := range latencies {
		sum += d
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}
	avg := sum / time.Duration(concurrency)
	p99Idx := int(float64(concurrency) * 0.99)
	// 简易排序取 P99
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	p99 := sorted[p99Idx]

	fmt.Printf("  %-20s | 并发=%d | avg=%-10s | P99=%-10s | max=%-10s | 被阻塞=%d/%d\n",
		name, concurrency, avg.Round(time.Microsecond), p99.Round(time.Microsecond),
		max.Round(time.Microsecond), totalBlocked, concurrency)
}

func main() {
	fmt.Println("=== 缓存击穿方案延迟对比 ===")
	fmt.Println("模拟条件：DB 查询延迟 50ms，所有请求同时到达同一个 hot_key")
	fmt.Println()

	_ = rand.Int // suppress unused import

	concurrencies := []int{100, 500, 1000}

	for _, c := range concurrencies {
		fmt.Printf("--- 并发 %d ---\n", c)

		// 方案 1：互斥锁
		mu1 := &sync.Mutex{}
		cache1 := make(map[string]string)
		runConcurrentTest("互斥锁(Mutex)", c, func(key string) time.Duration {
			_, d := withMutex(key, mu1, cache1)
			return d
		})

		// 方案 2：singleflight
		g := &singleflight.Group{}
		runConcurrentTest("singleflight", c, func(key string) time.Duration {
			_, d := withSingleflight(key, g)
			return d
		})

		// 方案 3：逻辑过期
		lc := newLogicalExpireCache()
		// 预热：先写入一条过期数据
		lc.mu.Lock()
		lc.data["hot_key"] = "stale_value"
		lc.expireAt["hot_key"] = time.Now().Add(-1 * time.Second) // 已过期
		lc.mu.Unlock()
		runConcurrentTest("逻辑过期", c, func(key string) time.Duration {
			_, d := lc.get(key)
			return d
		})

		fmt.Println()
	}

	fmt.Println("=== 关键发现 ===")
	fmt.Println("1. 互斥锁：所有请求串行等待，延迟随并发线性增长")
	fmt.Println("2. singleflight：只有 1 个请求回源，其余共享结果，延迟≈1次DB查询")
	fmt.Println("3. 逻辑过期：返回旧值不阻塞，延迟极低，但数据一致性牺牲")
}
