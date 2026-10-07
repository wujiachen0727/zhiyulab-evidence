package main

import "fmt"

// scenario1 参数在 defer 语句处就求值，闭包读变量则推迟到执行时
func scenario1() {
	fmt.Println("== 1. 参数求值时机 ==")

	x := 1
	defer fmt.Println("  defer 参数打印 x =", x) // 参数此刻就固定为 1
	x = 100
	fmt.Println("  函数退出前 x =", x)

	y := 1
	defer func() { fmt.Println("  defer 闭包读到 y =", y) }() // 执行时才读 y
	y = 200
	fmt.Println("  函数退出前 y =", y)
}

// scenario2 命名返回值可以被 defer 改写，普通返回值不行
func plainReturn() int {
	r := 50
	defer func() { r = 100 }()
	return r
}

func namedReturn() (r int) {
	defer func() { r = 100 }()
	return 50
}

func scenario2() {
	fmt.Println("== 2. 返回值能不能被 defer 改 ==")
	fmt.Printf("  普通返回值 return r -> %d\n", plainReturn())
	fmt.Printf("  命名返回值 return 50 -> %d\n", namedReturn())

	s := []int{}
	defer func() { fmt.Println("  函数退出时 s =", s) }()
	s = append(s, 1)
	fmt.Println("  append 之后 s =", s)
}

// scenario3 循环里的 defer 到函数结束才执行，顺序是后进先出
func scenario3() {
	fmt.Println("== 3. 循环里 defer 的执行时机与顺序 ==")
	for i := 1; i <= 3; i++ {
		defer fmt.Printf("  第 %d 次迭代注册的 defer 执行\n", i)
	}
	fmt.Println("  循环已经跑完了，下面才轮到 defer")
}

func main() {
	scenario1()
	fmt.Println()
	scenario2()
	fmt.Println()
	scenario3()
	fmt.Println()
	fmt.Println("== 4. 循环里 defer 的堆分配（见 output/ 的汇编检查）==")
}
