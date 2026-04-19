package main

import (
	"testing"
)

// ========== 内联友好版本 ==========

// 简单小函数，编译器可以内联
func add(a, b int) int {
	return a + b
}

// 调用内联友好函数的循环
func callAdd() int {
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += add(i, i+1)
	}
	return sum
}

// ========== 接口调用版本（编译器不可内联） ==========

type Calculator interface {
	Calculate(a, b int) int
}

type concreteAdder struct{}

func (c concreteAdder) Calculate(a, b int) int {
	return a + b
}

// 通过动态分发确保编译器无法去虚化
// 关键：只在运行时才知道具体类型
func dynamicInterfaceCall() int {
	var calc Calculator
	// 条件分支让编译器无法在编译期确定具体类型
	// 这样接口调用必须动态分发，无法内联
	calcs := []Calculator{concreteAdder{}}
	calc = calcs[0]
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += calc.Calculate(i, i+1)
	}
	return sum
}

// ========== Benchmark ==========

func BenchmarkDirectCall_InlineFriendly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		callAdd()
	}
}

func BenchmarkInterfaceCall_NoInline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dynamicInterfaceCall()
	}
}
