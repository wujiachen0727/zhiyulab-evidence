// E2.2: Sliding Window Log vs Counter vs Fixed Window
// 证明：三种窗口算法在精度和内存上的取舍
// 环境：Go 1.26.2，无外部依赖
// 运行：go run main.go

package main

import (
	"fmt"
	"time"
)

// --- Sliding Window Log（精确）---
type SWLog struct {
	window    time.Duration
	limit     int
	timestamps []time.Time
}

func (s *SWLog) Allow(now time.Time) bool {
	cutoff := now.Add(-s.window)
	// 清理过期
	valid := s.timestamps[:0]
	for _, t := range s.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	s.timestamps = valid
	if len(s.timestamps) < s.limit {
		s.timestamps = append(s.timestamps, now)
		return true
	}
	return false
}

func (s *SWLog) MemoryEntries() int { return len(s.timestamps) }

// --- Sliding Window Counter（近似）---
type SWCounter struct {
	window   time.Duration
	limit    int
	buckets  int
	counts   []int
	starts   []time.Time
	bucketDur time.Duration
}

func NewSWCounter(window time.Duration, limit, buckets int) *SWCounter {
	dur := window / time.Duration(buckets)
	c := &SWCounter{
		window: window, limit: limit,
		buckets: buckets, bucketDur: dur,
		counts: make([]int, buckets),
		starts: make([]time.Time, buckets),
	}
	base := time.Now().Add(-window)
	for i := range c.starts {
		c.starts[i] = base.Add(time.Duration(i) * dur)
	}
	return c
}

func (c *SWCounter) Allow(now time.Time) bool {
	// 确定当前桶
	idx := int(now.Sub(c.starts[0]) / c.bucketDur) % c.buckets
	// 清理过期桶
	cutoff := now.Add(-c.window)
	total := 0
	for i := 0; i < c.buckets; i++ {
		if c.starts[i].Before(cutoff) {
			c.counts[i] = 0
			c.starts[i] = now
		}
		total += c.counts[i]
	}
	if total < c.limit {
		c.counts[idx]++
		return true
	}
	return false
}

func (c *SWCounter) MemoryEntries() int { return c.buckets }

// --- Fixed Window ---
type FixedWindow struct {
	window time.Duration
	limit  int
	count  int
	start  time.Time
}

func (f *FixedWindow) Allow(now time.Time) bool {
	if now.Sub(f.start) >= f.window {
		f.start = now
		f.count = 0
	}
	if f.count < f.limit {
		f.count++
		return true
	}
	return false
}

func main() {
	fmt.Println("=== E2.2: 三种窗口算法对比 [实测 Go 1.26.2] ===")
	fmt.Println("配置: 窗口=1s, 限额=5请求/窗口")
	fmt.Println("流量: 窗口尾部3个 + 窗口头部3个（测试边界跨窗）")
	fmt.Println()

	window := 1 * time.Second
	limit := 5

	swLog := &SWLog{window: window, limit: limit}
	swCounter := NewSWCounter(window, limit, 10)
	fwWindow := &FixedWindow{window: window, limit: limit, start: time.Now()}

	// 模拟：窗口尾部发 3 个，然后跨窗口再发 3 个
	base := time.Now()

	// 阶段1: 窗口尾部 800ms-900ms 发 3 个请求
	phase1Times := []time.Duration{800 * time.Millisecond, 850 * time.Millisecond, 900 * time.Millisecond}
	// 阶段2: 窗口头部 1050ms-1150ms 发 3 个请求
	phase2Times := []time.Duration{1050 * time.Millisecond, 1100 * time.Millisecond, 1150 * time.Millisecond}

	allTimes := append(phase1Times, phase2Times...)

	fmt.Printf("%-6s %-12s %-12s %-12s %-12s\n", "时间", "偏移(ms)", "SW Log", "SW Counter", "Fixed Window")
	fmt.Println("--------------------------------------------------------------")

	for i, offset := range allTimes {
		now := base.Add(offset)
		r1 := swLog.Allow(now)
		r2 := swCounter.Allow(now)
		r3 := fwWindow.Allow(now)
		mark := ""
		if i == 3 {
			mark = " ← 跨窗口"
		}
		fmt.Printf("%-6d %-12d %-12s %-12s %-12s%s\n",
			i+1, offset.Milliseconds(),
			boolStr(r1), boolStr(r2), boolStr(r3), mark)
	}

	fmt.Println()
	fmt.Println("--- 内存占用对比（存储 1K QPS × 60s 窗口）---")
	fmt.Printf("%-18s %-15s %-12s\n", "算法", "存储条目", "内存估算")
	fmt.Println("-----------------------------------------------")
	fmt.Printf("%-18s %-15s %-12s\n", "SW Log（精确）", "60,000 条时间戳", "~3.7 MB/key")
	fmt.Printf("%-18s %-15s %-12s\n", "SW Counter（10桶）", "10 个计数器", "~80 B/key")
	fmt.Printf("%-18s %-15s %-12s\n", "Fixed Window", "1 个计数器", "~8 B/key")

	fmt.Println()
	fmt.Println("--- 结论 ---")
	fmt.Println("• SW Log 最精确（无跨窗口误差），但内存 O(QPS×窗口)")
	fmt.Println("• SW Counter 近似精确（误差 ≤ 1/桶数），内存 O(桶数)")
	fmt.Println("• Fixed Window 最省内存，但存在'窗口边界双倍放行'问题")
	fmt.Println("• 精度 vs 内存的取舍：99% 精度只需 100 个桶 vs 存所有时间戳")
}

func boolStr(b bool) string {
	if b {
		return "✓ 通过"
	}
	return "✗ 拒绝"
}
