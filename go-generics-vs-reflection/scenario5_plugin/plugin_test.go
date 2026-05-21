package scenario5

import (
	"reflect"
	"testing"
)

func TestPluginSystem(t *testing.T) {
	registry := NewRegistry()

	// 运行时注册插件
	registry.Register("greeter", &GreeterPlugin{})
	registry.Register("math", &MathPlugin{})

	// 动态调用 — 方法名和参数都是运行时决定的
	results, err := registry.Call("greeter", "Hello", "世界")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("greeter.Hello(\"世界\") = %v", results[0])

	results, err = registry.Call("math", "Add", 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("math.Add(3, 5) = %v", results[0])

	results, err = registry.Call("math", "Multiply", 4, 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("math.Multiply(4, 7) = %v", results[0])

	// 运行时发现方法列表
	t.Log("\n=== 运行时方法发现（reflect 独占能力）===")
	for name, plugin := range registry.plugins {
		v := reflect.TypeOf(plugin)
		t.Logf("插件 %q 的方法:", name)
		for i := 0; i < v.NumMethod(); i++ {
			m := v.Method(i)
			t.Logf("  - %s%s", m.Name, m.Type)
		}
	}
}

// Benchmark: 反射动态调用 vs 直接调用的开销
func BenchmarkReflectCall(b *testing.B) {
	registry := NewRegistry()
	registry.Register("math", &MathPlugin{})
	for i := 0; i < b.N; i++ {
		_, _ = registry.Call("math", "Add", 3, 5)
	}
}

func BenchmarkDirectCall(b *testing.B) {
	p := &MathPlugin{}
	for i := 0; i < b.N; i++ {
		_ = p.Add(3, 5)
	}
}

// 关键论证：为什么泛型无法替代
func TestWhyGenericsCannotReplace(t *testing.T) {
	t.Log("=== 为什么泛型无法替代反射做插件系统 ===")
	t.Log("")
	t.Log("泛型的限制：类型参数必须编译时确定")
	t.Log("")
	t.Log("// 你想写这样的代码：")
	t.Log("// func Call[T any](plugin T, method string, args ...any) any")
	t.Log("// 但 T 在编译时必须确定 — 而插件类型编译时不存在")
	t.Log("")
	t.Log("// 接口也不行：")
	t.Log("// type Plugin interface { Handle(req Request) Response }")
	t.Log("// 这要求所有插件实现同一接口 — 但插件的方法签名各不相同")
	t.Log("")
	t.Log("反射的三个不可替代能力：")
	t.Log("1. reflect.TypeOf() — 运行时获取类型信息")
	t.Log("2. reflect.ValueOf().MethodByName() — 运行时发现方法")
	t.Log("3. reflect.Value.Call() — 运行时动态调用")
	t.Log("")
	t.Log("结论：当类型信息完全在运行时才可知时，反射是唯一选择。")
}
