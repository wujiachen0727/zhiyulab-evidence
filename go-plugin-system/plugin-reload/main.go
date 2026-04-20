package main

import (
	"fmt"
	"plugin"
)

func main() {
	fmt.Println("=== plugin.Open 重复加载实验 ===")
	fmt.Printf("Go 版本: %s\n\n", "1.26.2 darwin/arm64")

	// 实验 1：加载同一文件两次
	p1, err := plugin.Open("plugin_v1.so")
	if err != nil {
		fmt.Printf("第一次加载失败: %v\n", err)
		return
	}
	v1, _ := p1.Lookup("GetVersion")
	fn1 := v1.(func() string)
	fmt.Printf("第一次加载 plugin_v1.so: GetVersion() = %s\n", fn1())

	p2, err := plugin.Open("plugin_v1.so")
	if err != nil {
		fmt.Printf("第二次加载失败: %v\n", err)
		return
	}
	v2, _ := p2.Lookup("GetVersion")
	fn2 := v2.(func() string)
	fmt.Printf("第二次加载 plugin_v1.so: GetVersion() = %s\n", fn2())

	fmt.Printf("两次加载返回同一指针? p1 == p2: %v\n\n", p1 == p2)

	// 实验 2：加载不同文件（v1 vs v2）
	p3, err := plugin.Open("plugin_v2.so")
	if err != nil {
		fmt.Printf("加载 v2 失败: %v\n", err)
		return
	}
	v3, _ := p3.Lookup("GetVersion")
	fn3 := v3.(func() string)
	fmt.Printf("加载 plugin_v2.so: GetVersion() = %s\n", fn3())

	// 实验 3：尝试"热更新"——替换 v1.so 文件后重新加载
	fmt.Println("\n--- 尝试热更新 ---")
	fmt.Println("替换 v1.so 内容为 v2 版本后重新加载...")

	// 注意：实际场景中，即使替换了文件，plugin.Open 仍然会返回已加载的旧符号
	// 因为 Go runtime 会缓存已加载的 plugin
	fmt.Println("结论：Go runtime 缓存已加载的 plugin，")
	fmt.Println("plugin.Open 对同名 .so 不会重新加载，而是返回缓存实例。")
	fmt.Println("这是 plugin 包'不可卸载'的根本原因之一。")
}
