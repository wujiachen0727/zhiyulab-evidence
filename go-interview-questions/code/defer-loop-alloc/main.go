package main

import "fmt"

// noop 用 noinline 保证不被内联，让 defer 的目标函数保持真实调用
//
//go:noinline
func noop() {}

// nonLoop 的 defer 不在循环里，编译器可以用开放编码把 defer 记录放在栈上
func nonLoop() {
	defer noop()
	noop()
}

// inLoop 的 defer 在循环里，编译器无法在编译期确定迭代次数，只能走堆分配
func inLoop(n int) {
	for i := 0; i < n; i++ {
		defer noop()
	}
}

func main() {
	nonLoop()
	inLoop(2)
	fmt.Println("两个函数都执行完毕")
}
