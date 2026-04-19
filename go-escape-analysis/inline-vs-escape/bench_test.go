package main

import "testing"

// 被测函数：纯值类型小函数
func add(a, b int) int {
	return a + b
}

// 被测函数：带指针的小函数
func doubleVal(val *int) int {
	*val = *val * 2
	return *val
}

// 内联开启时的 benchmark
// go test -bench=. -benchmem -gcflags='-m'

var sink int

func BenchmarkAdd_InlineEnabled(b *testing.B) {
	x, y := 10, 20
	for i := 0; i < b.N; i++ {
		sink = add(x, y)
	}
}

func BenchmarkDoubleVal_InlineEnabled(b *testing.B) {
	val := 10
	for i := 0; i < b.N; i++ {
		sink = doubleVal(&val)
	}
}

func BenchmarkAddDirect(b *testing.B) {
	x, y := 10, 20
	for i := 0; i < b.N; i++ {
		sink = x + y // 直接计算，无函数调用
	}
}
