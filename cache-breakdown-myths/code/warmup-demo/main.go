package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CacheWarmer 最小可运行的缓存预热脚本
// 展示一个"看似简单"的预热需求的实际复杂度
type CacheWarmer struct {
	batchSize    int
	concurrency  int
	retryMax     int
	retryBackoff time.Duration
	stats        struct {
		total    int
		success  int
		failed   int
		skipped  int
		duration time.Duration
	}
}

func NewCacheWarmer() *CacheWarmer {
	return &CacheWarmer{
		batchSize:    500,
		concurrency:  10,
		retryMax:     3,
		retryBackoff: 100 * time.Millisecond,
	}
}

// 模拟：从数据库扫描需要预热的 key
func (w *CacheWarmer) scanKeys(ctx context.Context) ([]string, error) {
	keys := make([]string, 10000) // 模拟 1 万个 key
	for i := range keys {
		keys[i] = fmt.Sprintf("product:%d", i+1)
	}
	return keys, nil
}

// 模拟：从 DB 查询并写入缓存
func (w *CacheWarmer) warmKey(ctx context.Context, key string) error {
	// 模拟 DB 查询延迟 1-5ms
	time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)

	// 模拟 5% 的随机失败
	if rand.Float64() < 0.05 {
		return errors.New("db connection timeout")
	}

	// 模拟写入缓存
	time.Sleep(time.Duration(rand.Intn(2)) * time.Millisecond)
	return nil
}

// 带重试的预热
func (w *CacheWarmer) warmKeyWithRetry(ctx context.Context, key string) error {
	var lastErr error
	for attempt := 0; attempt <= w.retryMax; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}
		if err := w.warmKey(ctx, key); err != nil {
			lastErr = err
			// 指数退避
			backoff := w.retryBackoff * time.Duration(1<<uint(attempt))
			time.Sleep(backoff)
			continue
		}
		return nil
	}
	return fmt.Errorf("after %d retries: %w", w.retryMax, lastErr)
}

// 分批并发预热
func (w *CacheWarmer) Run(ctx context.Context) error {
	start := time.Now()

	keys, err := w.scanKeys(ctx)
	if err != nil {
		return fmt.Errorf("scan keys: %w", err)
	}
	w.stats.total = len(keys)

	fmt.Printf("预热开始：%d 个 key，批大小 %d，并发 %d\n", len(keys), w.batchSize, w.concurrency)

	var mu sync.Mutex
	sem := make(chan struct{}, w.concurrency)

	for batchStart := 0; batchStart < len(keys); batchStart += w.batchSize {
		batchEnd := batchStart + w.batchSize
		if batchEnd > len(keys) {
			batchEnd = len(keys)
		}
		batch := keys[batchStart:batchEnd]

		var wg sync.WaitGroup
		for _, key := range batch {
			wg.Add(1)
			sem <- struct{}{} // 限流
			go func(k string) {
				defer wg.Done()
				defer func() { <-sem }()

				if err := w.warmKeyWithRetry(ctx, k); err != nil {
					mu.Lock()
					w.stats.failed++
					mu.Unlock()
				} else {
					mu.Lock()
					w.stats.success++
					mu.Unlock()
				}
			}(key)
		}
		wg.Wait()

		// 批次间报告进度
		fmt.Printf("  已完成 %d/%d (成功=%d, 失败=%d)\n",
			batchEnd, len(keys), w.stats.success, w.stats.failed)
	}

	w.stats.duration = time.Since(start)

	fmt.Printf("\n=== 预热完成 ===\n")
	fmt.Printf("  总计：%d 个 key\n", w.stats.total)
	fmt.Printf("  成功：%d\n", w.stats.success)
	fmt.Printf("  失败：%d\n", w.stats.failed)
	fmt.Printf("  耗时：%s\n", w.stats.duration.Round(time.Millisecond))
	fmt.Printf("  吞吐：%.0f keys/sec\n", float64(w.stats.total)/w.stats.duration.Seconds())

	return nil
}

func main() {
	fmt.Println("=== 缓存预热脚本 Demo ===")
	fmt.Println("展示一个\"看似简单\"的预热需求的实际复杂度")
	fmt.Println()

	warmer := NewCacheWarmer()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := warmer.Run(ctx); err != nil {
		fmt.Printf("预热失败：%v\n", err)
	}

	fmt.Println()
	fmt.Println("=== 维护点清单（这些面试不会考你） ===")
	fmt.Println("1. 数据源遍历：游标分页 vs 全量扫描？大表怎么办？")
	fmt.Println("2. 分批控制：batch size 多大合适？太大打爆 DB，太小预热太慢")
	fmt.Println("3. 并发限流：sem channel 限流够用吗？需要令牌桶吗？")
	fmt.Println("4. 错误处理：单 key 失败要不要中止整个预热？重试几次？退避策略？")
	fmt.Println("5. 幂等性：预热到一半中断，重跑会覆盖刚写入的新数据吗？")
	fmt.Println("6. 监控：怎么知道预热是否完成？成功率多少算合格？")
	fmt.Println("7. 触发时机：部署时跑？定时跑？缓存重建时跑？")
	fmt.Println("8. 资源隔离：预热流量和线上流量走同一个 DB 连接池吗？")

	fmt.Println()
	fmt.Printf("以上 Demo 代码量：~120 行 Go 代码\n")
	fmt.Println("生产版本通常需要 300-500 行（加上配置、监控、日志、优雅停止）")
}
