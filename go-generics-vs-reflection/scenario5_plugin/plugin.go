// 场景5：插件系统/动态分发 —— 证明反射不可替代
package scenario5

import (
	"fmt"
	"reflect"
)

// === 插件注册中心（反射版）===
// 核心问题：编译时不知道将来会有什么插件

type PluginRegistry struct {
	plugins map[string]interface{} // 运行时注册，编译时类型未知
}

func NewRegistry() *PluginRegistry {
	return &PluginRegistry{plugins: make(map[string]interface{})}
}

// 注册插件（编译时不知道 plugin 的具体类型）
func (r *PluginRegistry) Register(name string, plugin interface{}) {
	r.plugins[name] = plugin
}

// 动态调用插件方法（只有反射能做到）
func (r *PluginRegistry) Call(pluginName, method string, args ...interface{}) ([]interface{}, error) {
	plugin, ok := r.plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", pluginName)
	}

	v := reflect.ValueOf(plugin)
	m := v.MethodByName(method)
	if !m.IsValid() {
		return nil, fmt.Errorf("method %q not found in plugin %q", method, pluginName)
	}

	// 构造参数
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 动态调用
	results := m.Call(in)

	// 提取返回值
	out := make([]interface{}, len(results))
	for i, r := range results {
		out[i] = r.Interface()
	}
	return out, nil
}

// === 示例插件（编译时未知——可能是第三方开发的） ===

type GreeterPlugin struct{}

func (p *GreeterPlugin) Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func (p *GreeterPlugin) Version() string {
	return "1.0.0"
}

type MathPlugin struct{}

func (p *MathPlugin) Add(a, b int) int {
	return a + b
}

func (p *MathPlugin) Multiply(a, b int) int {
	return a * b
}

// === 为什么泛型做不到？===
//
// 泛型需要编译时确定类型参数。但插件系统的核心需求是：
// 1. 运行时注册（插件可能是动态加载的 .so）
// 2. 运行时发现方法（不知道插件有哪些方法）
// 3. 运行时动态调用（参数类型运行时才知道）
//
// 你能写 func Call[T Plugin](p T, method string) 吗？
// 不能——因为 T 必须在编译时确定，而插件的类型编译时根本不存在。
//
// 这就是反射的不可替代领地：
// 当类型信息完全在运行时才可知时，泛型无能为力。
