package main

import (
	"testing"
)

// ========== 优化链完整版（内联→逃逸消除→死代码消除） ==========

// 小函数，编译器可以内联
// 内联后：调用点展开，编译器看到完整上下文
// 逃逸分析：内联后发现 sum 不逃逸 → 栈分配
// 死代码消除：内联后发现 unused 分支永不执行 → 消除
func addInline(a, b int) int {
	return a + b
}

//go:noinline
func chainComplete() int {
	sum := 0
	for i := 0; i < 1000; i++ {
		// 内联后编译器看到 addInline(i, i+1) = i + (i+1)
		// 进一步优化：可以消除循环中的函数调用开销
		sum += addInline(i, i+1)
	}
	return sum
}

// ========== 优化链断裂版（内联被阻断→逃逸无法优化→连锁失效） ==========

// 同样逻辑但通过接口调用，阻断内联
type Adder interface {
	Add(a, b int) int
}

type inlineAdder struct{}

func (a inlineAdder) Add(x, y int) int {
	return x + y
}

// 从 slice 取接口值，阻断去虚化，进而阻断内联
// 内联断裂 → 编译器看不到函数体 → 无法做逃逸优化和死代码消除
//go:noinline
func chainBroken() int {
	adders := []Adder{inlineAdder{}}
	a := adders[0]  // 动态分发，无法去虚化，无法内联
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += a.Add(i, i+1)  // 每次循环都要做接口分发
	}
	return sum
}

// ========== 对照组：纯直接计算（无函数调用） ==========

//go:noinline
func pureCalc() int {
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i + (i + 1)  // 这就是内联后编译器"看到"的代码
	}
	return sum
}

// ========== Benchmark ==========

func BenchmarkChain_Complete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		chainComplete()
	}
}

func BenchmarkChain_Broken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		chainBroken()
	}
}

func BenchmarkChain_PureCalc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pureCalc()
	}
}
