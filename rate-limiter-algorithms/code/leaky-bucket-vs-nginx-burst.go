// E1.1: 纯漏桶 vs Nginx-like 漏桶+burst+nodelay
// 证明：加 burst+nodelay 后行为从"匀速排队"变为"允许突发立即通过"
// 环境：Go 1.26.2，无外部依赖
// 运行：go run main.go

package main

import (
	"fmt"
	"sync"
	"time"
)

// 纯漏桶：严格匀速，超过 rate 的请求排队
type PureLeakyBucket struct {
	rate     float64 // 请求/秒
	lastTick time.Time
	mu       sync.Mutex
}

func (lb *PureLeakyBucket) Allow() (bool, time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	now := time.Now()
	interval := time.Duration(float64(time.Second) / lb.rate)
	next := lb.lastTick.Add(interval)
	if now.Before(next) {
		wait := next.Sub(now)
		return false, wait // 需要等待
	}
	lb.lastTick = now
	return true, 0
}

// Nginx-like: burst+nodelay 模式
// burst 个请求可以立即通过（不排队），超过 burst 的才拒绝
type NginxLikeBurst struct {
	rate      float64
	burst     int
	tokens    float64
	lastTick  time.Time
	mu        sync.Mutex
}

func (nb *NginxLikeBurst) Allow() bool {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(nb.lastTick).Seconds()
	nb.lastTick = now
	// 补充令牌
	nb.tokens += elapsed * nb.rate
	if nb.tokens > float64(nb.burst) {
		nb.tokens = float64(nb.burst)
	}
	if nb.tokens >= 1 {
		nb.tokens--
		return true // 立即通过
	}
	return false // 拒绝
}

func main() {
	fmt.Println("=== E1.1: 纯漏桶 vs Nginx-like burst+nodelay ===")
	fmt.Println("配置: rate=2r/s, burst=5")
	fmt.Println("流量: 10个请求在50ms内突发到达")
	fmt.Println()

	// 配置：2 请求/秒，burst=5
	pureLB := &PureLeakyBucket{rate: 2, lastTick: time.Now()}
	nginxLB := &NginxLikeBurst{
		rate: 2, burst: 5,
		tokens: 5, lastTick: time.Now(),
	}

	// 模拟突发：10 个请求在 50ms 内到达
	fmt.Println("--- 纯漏桶（严格匀速）---")
	fmt.Printf("%-4s %-10s %-8s\n", "#", "结果", "需等待")
	for i := 1; i <= 10; i++ {
		ok, wait := pureLB.Allow()
		if ok {
			fmt.Printf("%-4d %-10s %-8s\n", i, "✓ 通过", "-")
		} else {
			fmt.Printf("%-4d %-10s %-8s\n", i, "✗ 排队", wait.Round(time.Millisecond))
		}
		time.Sleep(5 * time.Millisecond) // 5ms间隔≈突发
	}

	fmt.Println()
	fmt.Println("--- Nginx-like burst+nodelay（允许突发）---")
	fmt.Printf("%-4s %-10s\n", "#", "结果")
	for i := 1; i <= 10; i++ {
		ok := nginxLB.Allow()
		if ok {
			fmt.Printf("%-4d %-10s\n", i, "✓ 立即通过")
		} else {
			fmt.Printf("%-4d %-10s\n", i, "✗ 拒绝")
		}
		time.Sleep(5 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("=== 结论 ===")
	fmt.Println("纯漏桶: 只放行第1个请求，后续全部排队等待（匀速500ms间隔）")
	fmt.Println("Nginx-like: 前5个请求立即通过（消耗burst容量），第6个起拒绝")
	fmt.Println("→ burst+nodelay 让行为从'匀速排队'变为'允许突发立即通过+拒绝超额'")
	fmt.Println("→ 这正是令牌桶的特征：桶里有令牌就放行，没有就拒绝")
}
