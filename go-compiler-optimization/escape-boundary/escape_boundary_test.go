package main

import (
	"testing"
)

// ========== 防止内联的逃逸对照 ==========

//go:noinline
func returnByValue() int {
	x := 42
	return x
}

//go:noinline
func returnByPointer() *int {
	x := 42
	return &x
}

//go:noinline
func closureCapture() func() int {
	x := 42
	return func() int {
		return x
	}
}

//go:noinline
func interfaceConvert() interface{} {
	x := 42
	return x
}

//go:noinline
func sliceAlloc() []int {
	s := make([]int, 0, 1)
	s = append(s, 42)
	s = append(s, 43)
	return s
}

// ========== 确保返回值被使用 ==========

var sink interface{}

// ========== Benchmark ==========

func BenchmarkStackAlloc_NoEscape(b *testing.B) {
	for i := 0; i < b.N; i++ {
		returnByValue()
	}
}

func BenchmarkHeapAlloc_ReturnPointer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = returnByPointer()
	}
}

func BenchmarkHeapAlloc_ClosureCapture(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = closureCapture()
	}
}

func BenchmarkHeapAlloc_InterfaceConvert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = interfaceConvert()
	}
}

func BenchmarkHeapAlloc_SliceAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = sliceAlloc()
	}
}
