package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/redis/go-redis/v9"
)

// E3 补测：本地 bigcache GET vs Docker Redis GET 延迟对比
// 同一台 M4 Pro，Redis 走 Docker localhost TCP

func main() {
	ctx := context.Background()
	const numKeys = 10_000
	const numOps = 10_000
	val := make([]byte, 64)
	rand.Read(val)

	// 初始化 bigcache
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.Verbose = false
	bc, _ := bigcache.New(ctx, config)
	for i := 0; i < numKeys; i++ {
		bc.Set(strconv.Itoa(i), val)
	}

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("Redis 连接失败:", err)
		return
	}
	for i := 0; i < numKeys; i++ {
		rdb.Set(ctx, strconv.Itoa(i), val, 0)
	}

	// 测试 bigcache GET 延迟
	fmt.Println("=== bigcache GET 延迟 (10000 次) ===")
	bcLatencies := make([]time.Duration, numOps)
	for i := 0; i < numOps; i++ {
		key := strconv.Itoa(rand.Intn(numKeys))
		start := time.Now()
		bc.Get(key)
		bcLatencies[i] = time.Since(start)
	}
	printStats("bigcache", bcLatencies)

	// 测试 Redis GET 延迟
	fmt.Println("\n=== Redis GET 延迟 (10000 次, Docker localhost) ===")
	redisLatencies := make([]time.Duration, numOps)
	for i := 0; i < numOps; i++ {
		key := strconv.Itoa(rand.Intn(numKeys))
		start := time.Now()
		rdb.Get(ctx, key)
		redisLatencies[i] = time.Since(start)
	}
	printStats("Redis", redisLatencies)

	// 清理
	rdb.FlushAll(ctx)
	rdb.Close()
	bc.Close()
}

func printStats(name string, latencies []time.Duration) {
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	n := len(latencies)
	p50 := latencies[n*50/100]
	p90 := latencies[n*90/100]
	p99 := latencies[n*99/100]

	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	avg := total / time.Duration(n)

	fmt.Printf("%s:\n", name)
	fmt.Printf("  平均: %v\n", avg)
	fmt.Printf("  P50:  %v\n", p50)
	fmt.Printf("  P90:  %v\n", p90)
	fmt.Printf("  P99:  %v\n", p99)
}
