// 测试目标文件：包含短函数和超长函数，供 funclength linter 检测
package example

import "fmt"

// ShortFunc 只有 5 行，不会触发警告
func ShortFunc() {
	x := 1
	y := 2
	z := x + y
	fmt.Println(z)
}

// TooLongFunc 超过 80 行，会触发警告
func TooLongFunc() { // want `函数 TooLongFunc 有 \d+ 行，超过上限 80 行`
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
	fmt.Println(a + 1)
	fmt.Println(b + 1)
	fmt.Println(c + 1)
	fmt.Println(d + 1)
	fmt.Println(e + 1)
	fmt.Println(a * 2)
	fmt.Println(b * 2)
	fmt.Println(c * 2)
	fmt.Println(d * 2)
	fmt.Println(e * 2)
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
	_ = a * 3
	_ = b * 3
	_ = c * 3
	_ = d * 3
	_ = e * 3
	_ = a + b + c
	_ = b + c + d
	_ = c + d + e
	_ = a + b + c + d
	_ = b + c + d + e
	_ = a + b + c + d + e
	fmt.Println(a + 10)
	fmt.Println(b + 10)
	fmt.Println(c + 10)
	fmt.Println(d + 10)
	fmt.Println(e + 10)
	fmt.Println(a * 10)
	fmt.Println(b * 10)
	fmt.Println(c * 10)
	fmt.Println(d * 10)
	fmt.Println(e * 10)
	fmt.Println(a + b + 10)
	fmt.Println(b + c + 10)
	fmt.Println(c + d + 10)
	fmt.Println(d + e + 10)
	fmt.Println(a + e + 10)
	fmt.Println(a + b + c + 10)
	fmt.Println(b + c + d + 10)
	fmt.Println(c + d + e + 10)
	fmt.Println(a + b + c + d + 10)
	fmt.Println(b + c + d + e + 10)
	_ = a + 100
	_ = b + 100
	_ = c + 100
	_ = d + 100
	_ = e + 100
	fmt.Println("done")
}
