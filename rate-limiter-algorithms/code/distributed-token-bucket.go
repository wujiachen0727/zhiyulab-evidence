// E3.1: 朴素分布式令牌桶 vs Lua 原子版 并发超发实验
// 证明：非原子的 GET/SET 两步操作在并发下产生超发
// 环境：Go 1.26.2，无外部依赖（用本地 mock 模拟 Redis 竞态）
// 运行：go run main.go

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 模拟 Redis 存储（本地 mock）
type MockRedis struct {
	mu     sync.Mutex
	tokens float64
	lastTs time.Time
}

// --- 朴素版：GET → 计算 → SET（两步非原子）---
func (r *MockRedis) NaiveAllow(rate float64, capacity float64) bool {
	// Step 1: GET（读取当前令牌数）
	r.mu.Lock()
	currentTokens := r.tokens
	lastTs := r.lastTs
	r.mu.Unlock()

	// 模拟网络延迟（释放锁后其他 goroutine 可并发读）
	time.Sleep(100 * time.Microsecond)

	// Step 2: 计算新令牌
	now := time.Now()
	elapsed := now.Sub(lastTs).Seconds()
	newTokens := currentTokens + elapsed*rate
	if newTokens > capacity {
		newTokens = capacity
	}

	// Step 3: SET（扣减并写回）
	if newTokens >= 1 {
		r.mu.Lock()
		// 关键问题：多个 goroutine 都读到了"有令牌"，都来 SET
		r.tokens = newTokens - 1
		r.lastTs = now
		r.mu.Unlock()
		return true
	}
	return false
}

// --- Lua 原子版：整个操作在锁内（模拟 Redis Lua 脚本）---
func (r *MockRedis) AtomicAllow(rate float64, capacity float64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastTs).Seconds()
	r.tokens += elapsed * rate
	if r.tokens > capacity {
		r.tokens = capacity
	}
	r.lastTs = now

	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

func main() {
	fmt.Println("=== E3.1: 分布式令牌桶并发超发实验 [实测 Go 1.26.2] ===")
	fmt.Println("配置: rate=10令牌/s, capacity=10, 并发=100 goroutine")
	fmt.Println("预期：10个令牌最多放行10个请求")
	fmt.Println()

	const (
		goroutines = 100
		rate       = 10.0
		capacity   = 10.0
	)

	// 测试朴素版
	fmt.Println("--- 朴素版（GET/SET 两步，非原子）---")
	var naiveTotal int64
	for trial := 0; trial < 5; trial++ {
		redis := &MockRedis{tokens: capacity, lastTs: time.Now()}
		var passed int64
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				if redis.NaiveAllow(rate, capacity) {
					atomic.AddInt64(&passed, 1)
				}
			}()
		}
		wg.Wait()
		fmt.Printf("  第%d轮: 放行 %d / 100（超发 %d）\n",
			trial+1, passed, max(0, passed-int64(capacity)))
		naiveTotal += passed
	}
	naiveAvg := float64(naiveTotal) / 5
	fmt.Printf("  平均放行: %.1f / 100（预期最多 10）\n", naiveAvg)

	fmt.Println()

	// 测试原子版
	fmt.Println("--- Lua 原子版（整个操作原子化）---")
	var atomicTotal int64
	for trial := 0; trial < 5; trial++ {
		redis := &MockRedis{tokens: capacity, lastTs: time.Now()}
		var passed int64
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				if redis.AtomicAllow(rate, capacity) {
					atomic.AddInt64(&passed, 1)
				}
			}()
		}
		wg.Wait()
		fmt.Printf("  第%d轮: 放行 %d / 100（超发 %d）\n",
			trial+1, passed, max(0, passed-int64(capacity)))
		atomicTotal += passed
	}
	atomicAvg := float64(atomicTotal) / 5
	fmt.Printf("  平均放行: %.1f / 100（预期最多 10）\n", atomicAvg)

	fmt.Println()
	fmt.Println("=== 结论 ===")
	fmt.Printf("朴素版平均超发: %.0f 个请求（%.0f%%）\n",
		naiveAvg-capacity, (naiveAvg-capacity)/capacity*100)
	overAtom := atomicAvg - capacity
	if overAtom < 0 {
		overAtom = 0
	}
	fmt.Printf("原子版平均超发: %.0f 个请求\n", overAtom)
	fmt.Println("→ 非原子的 GET/SET 在并发下必然超发")
	fmt.Println("→ Lua 脚本/事务原子化是分布式令牌桶的硬性要求")
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
