package main

import (
	"testing"
)

// ========== BCE 消除版（编译器能证明索引安全） ==========

//go:noinline
func sumWithBCE(data []int) int {
	sum := 0
	// 先用 len 检查，编译器证明后续 data[i] 永远不会越界
	// prove pass 消除边界检查
	for i := 0; i < len(data); i++ {
		sum += data[i]
	}
	return sum
}

// ========== BCE 未消除版（编译器无法证明索引安全） ==========

//go:noinline
func sumWithoutBCE(data []int, indices []int) int {
	sum := 0
	// 用另一个 slice 的索引访问 data
	// 编译器无法证明 indices 中的值 < len(data)
	// 每次 data[indices[i]] 都要做边界检查
	for i := 0; i < len(indices); i++ {
		sum += data[indices[i]]
	}
	return sum
}

// ========== range 版本（天然有 BCE） ==========

//go:noinline
func sumWithRange(data []int) int {
	sum := 0
	// range 遍历，编译器自动证明索引安全
	for _, v := range data {
		sum += v
	}
	return sum
}

// ========== Benchmark ==========

func BenchmarkSum_WithBCE(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumWithBCE(data)
	}
}

func BenchmarkSum_WithoutBCE(b *testing.B) {
	data := make([]int, 1000)
	indices := make([]int, 1000)
	for i := range data {
		data[i] = i
		indices[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumWithoutBCE(data, indices)
	}
}

func BenchmarkSum_WithRange(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumWithRange(data)
	}
}
