package benchmark

import (
	"sync"
	"testing"
)

// ============================================================
// 场景4：管道 — Channel 多阶段处理 vs Mutex+slice
// 三阶段管道：generate → transform(×2) → aggregate
// Channel 连接各阶段 vs Mutex+slice 实现同样逻辑
// ============================================================

// ---------- Channel Pipeline：三阶段 ----------

func BenchmarkPipeline_Channel(b *testing.B) {
	stage1 := make(chan int, 64)
	stage2 := make(chan int, 64)
	done := make(chan int64, 1)

	// Stage 2: transform (×2)
	go func() {
		for v := range stage1 {
			stage2 <- v * 2
		}
		close(stage2)
	}()

	// Stage 3: aggregate (sum)
	go func() {
		var sum int64
		for v := range stage2 {
			sum += int64(v)
		}
		done <- sum
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stage1 <- i
	}
	close(stage1)
	<-done
}

// ---------- Mutex+slice Pipeline：同样三阶段逻辑 ----------

func BenchmarkPipeline_Mutex(b *testing.B) {
	var mu sync.Mutex
	data := make([]int, 0, 1024)
	transformed := make([]int, 0, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Stage 1: generate
		mu.Lock()
		data = append(data, i)
		mu.Unlock()
	}

	// Stage 2: transform
	mu.Lock()
	for _, v := range data {
		transformed = append(transformed, v*2)
	}
	mu.Unlock()

	// Stage 3: aggregate
	var sum int64
	mu.Lock()
	for _, v := range transformed {
		sum += int64(v)
	}
	mu.Unlock()
	_ = sum
}

// ---------- 顺序基线（无并发，无同步开销）----------

func BenchmarkPipeline_Sequential(b *testing.B) {
	var sum int64
	for i := 0; i < b.N; i++ {
		v := i * 2 // generate + transform 合并
		sum += int64(v)
	}
	_ = sum
}
