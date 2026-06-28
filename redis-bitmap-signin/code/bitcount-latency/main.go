// E5b：BITCOUNT 大 key 阻塞实测（扩展版）
// 在 E5 基础上增加 512MB（Redis 上限）和 redis-benchmark 对比
// 实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	sizes := []struct {
		name string
		bits int64
	}{
		{"1KB", 8 * 1024},
		{"1MB", 8 * 1024 * 1024},
		{"10MB", 80 * 1024 * 1024},
		{"100MB", 800 * 1024 * 1024},
		{"512MB", 4 * 1024 * 1024 * 1024}, // Redis Bitmap 上限
	}

	fmt.Println("=== E5b：BITCOUNT 大 key 耗时实测（扩展版） ===")
	fmt.Println("实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64")
	fmt.Println()
	fmt.Println("方法：对每个规模 Bitmap 跑 200 次 BITCOUNT，取平均和 P99")
	fmt.Println()

	for _, s := range sizes {
		key := fmt.Sprintf("sign:big:%s", s.name)

		start := time.Now()
		_ = rdb.SetBit(ctx, key, s.bits-1, 1).Err()
		constructTime := time.Since(start)

		strlen, _ := rdb.StrLen(ctx, key).Result()
		mem, _ := rdb.MemoryUsage(ctx, key).Result()

		fmt.Printf("--- %s Bitmap（STRLEN=%d 字节, MEMORY=%d 字节, 构造耗时=%v） ---\n", s.name, strlen, mem, constructTime)

		// 跑 200 次 BITCOUNT，记录每次耗时
		const iterations = 200
		times := make([]time.Duration, 0, iterations)
		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := rdb.BitCount(ctx, key, nil).Result()
			if err != nil {
				fmt.Printf("  BITCOUNT error: %v\n", err)
				continue
			}
			times = append(times, time.Since(start))
		}

		// 计算平均、P50、P95、P99
		var total time.Duration
		for _, t := range times {
			total += t
		}
		avg := total / time.Duration(len(times))
		p50 := percentile(times, 50)
		p95 := percentile(times, 95)
		p99 := percentile(times, 99)
		maxLatency := maxDuration(times)

		fmt.Printf("  BITCOUNT: avg=%v  P50=%v  P95=%v  P99=%v  max=%v\n", avg, p50, p95, p99, maxLatency)
		fmt.Println()
	}

	_ = os.Stdout.Sync()
}

func percentile(times []time.Duration, p int) time.Duration {
	if len(times) == 0 {
		return 0
	}
	// 简单排序
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	idx := len(sorted) * p / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func maxDuration(times []time.Duration) time.Duration {
	m := time.Duration(0)
	for _, t := range times {
		if t > m {
			m = t
		}
	}
	return m
}
