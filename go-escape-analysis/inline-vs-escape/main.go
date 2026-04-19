package main

import "fmt"

// 关键实验：纯值类型参数的小函数
// 内联开启时：函数边界消失，所有计算在栈上完成
// 内联关闭时：函数边界存在，但值类型参数仍在栈上
//
// 真正的差异在于：内联可以让后续逃逸分析看到完整上下文

// 小函数1：计算两点距离（纯值类型）
func distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

// 小函数2：创建一个 []int 并求和
// 注意：内联后，make 调用的上下文对逃逸分析可见
func createAndSum(n int) int {
	data := make([]int, n)
	sum := 0
	for i := range data {
		data[i] = i + 1
		sum += data[i]
	}
	return sum
}

func main() {
	d := distance(0, 0, 3, 4)
	fmt.Println("distance:", d)

	s := createAndSum(100)
	fmt.Println("sum:", s)
}
