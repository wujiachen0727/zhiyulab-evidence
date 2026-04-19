package main

import "fmt"

// 返回局部变量指针——经典逃逸场景
// 但如果函数被内联，编译器可能追踪到指针只在调用者内部使用
func createValue(n int) *int {
	v := n * 2
	return &v // 局部变量取地址返回
}

func main() {
	p := createValue(5)
	fmt.Println(*p)
}
