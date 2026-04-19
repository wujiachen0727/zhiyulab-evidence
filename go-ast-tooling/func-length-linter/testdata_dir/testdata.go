// 测试目标文件：包含长短函数，供 linter 检测
package testdata

import "fmt"

// shortFunc 只有 3 行，不会触发警告
func shortFunc() {
	x := 1
	y := 2
	fmt.Println(x + y)
}

// longFunc 超过 50 行，会触发警告
func longFunc() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
	fmt.Println(e)
	fmt.Println(a + b)
	fmt.Println(b + c)
	fmt.Println(c + d)
	fmt.Println(d + e)
	fmt.Println(a + e)
	fmt.Println(a * b)
	fmt.Println(b * c)
	fmt.Println(c * d)
	fmt.Println(d * e)
	fmt.Println(a * e)
	fmt.Println(a - b)
	fmt.Println(b - c)
	fmt.Println(c - d)
	fmt.Println(d - e)
	fmt.Println(a - e)
	fmt.Println(a / 1)
	fmt.Println(b / 1)
	fmt.Println(c / 1)
	fmt.Println(d / 1)
	fmt.Println(e / 1)
	fmt.Println(a + b + c)
	fmt.Println(b + c + d)
	fmt.Println(c + d + e)
	fmt.Println(a + b + c + d)
	fmt.Println(b + c + d + e)
	fmt.Println(a + b + c + d + e)
	_ = a + b
	_ = b + c
	_ = c + d
	_ = d + e
	_ = a + e
	_ = a * 2
	_ = b * 2
	_ = c * 2
	_ = d * 2
	_ = e * 2
	_ = a + 1
	_ = b + 1
	_ = c + 1
	_ = d + 1
	_ = e + 1
	_ = a + b + c + d + e
	fmt.Println("done")
}

// mediumFunc 30 行左右，不触发
func mediumFunc() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		fmt.Println(i * 2)
		fmt.Println(i * 3)
	}
	for j := 0; j < 5; j++ {
		fmt.Println(j)
		fmt.Println(j + 1)
		fmt.Println(j + 2)
	}
	fmt.Println("end")
}
