package main

import "fmt"

// 指针参数函数——内联的关键场景
// 内联开启：编译器看到 &x 只在 addOne 内部使用，不逃逸
// 内联关闭：编译器只看到函数接收 *int，保守逃逸
func addOne(n *int) int {
	*n = *n + 1
	return *n
}

func main() {
	x := 10
	result := addOne(&x)
	fmt.Println(result)
}
