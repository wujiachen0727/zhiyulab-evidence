package main

import (
	"testing"
)

type Processor interface {
	Process(data int) int
}

type fastProcessor struct{}

func (f fastProcessor) Process(data int) int {
	return data * 2
}

type slowProcessor struct{}

func (s slowProcessor) Process(data int) int {
	return data * 3
}

// ========== 场景1：编译器可去虚化 ==========
// 具体类型直接赋值，编译器能看到具体类型

//go:noinline
func concreteCall() int {
	var p Processor = fastProcessor{}
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += p.Process(i)
	}
	return sum
}

// ========== 场景2：编译器不可去虚化 ==========
// 通过 slice 间接获取，运行时才知道具体类型

//go:noinline
func dynamicCall() int {
	processors := []Processor{fastProcessor{}, slowProcessor{}}
	p := processors[0]  // 编译器不知道运行时取到哪个
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += p.Process(i)
	}
	return sum
}

// ========== 场景3：直接调用具体类型（对照基准） ==========

//go:noinline
func directCall() int {
	p := fastProcessor{}
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += p.Process(i)
	}
	return sum
}

// ========== Benchmark ==========

func BenchmarkConcreteInterface_Devirtualize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concreteCall()
	}
}

func BenchmarkDynamicInterface_NoDevirtualize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dynamicCall()
	}
}

func BenchmarkDirectCall_Baseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		directCall()
	}
}
