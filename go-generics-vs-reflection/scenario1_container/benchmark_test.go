package scenario1

import "testing"

// Benchmark: 反射版 Push + Pop 1000 次
func BenchmarkReflectStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := &ReflectStack{}
		for j := 0; j < 1000; j++ {
			s.Push(j)
		}
		for j := 0; j < 1000; j++ {
			val, _ := s.Pop()
			_ = val.(int) // type assertion — 运行时才知道类型
		}
	}
}

// Benchmark: 泛型版 Push + Pop 1000 次
func BenchmarkGenericStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := &GenericStack[int]{}
		for j := 0; j < 1000; j++ {
			s.Push(j)
		}
		for j := 0; j < 1000; j++ {
			val, _ := s.Pop()
			_ = val // 编译时已知是 int，无需断言
		}
	}
}

// 类型安全性对比测试
func TestTypeSafety(t *testing.T) {
	// 反射版：编译器不会阻止你 Push 错误类型
	rs := &ReflectStack{}
	rs.Push(42)
	rs.Push("oops") // 编译通过！运行时才爆炸

	val, _ := rs.Pop()
	// 如果你以为所有元素都是 int：
	// _ = val.(int)  // 这里会 panic: interface conversion

	t.Logf("反射版允许混入不同类型: %v (type: %T)", val, val)

	// 泛型版：编译时阻止
	gs := &GenericStack[int]{}
	gs.Push(42)
	// gs.Push("oops")  // 编译错误！cannot use "oops" (untyped string constant) as int value

	v, _ := gs.Pop()
	t.Logf("泛型版编译时保证类型: %d", v)
}
