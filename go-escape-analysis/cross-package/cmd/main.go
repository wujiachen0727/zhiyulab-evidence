package main

import (
	"fmt"
	"cross-package/internal"
)

// 包内函数：同包调用
func localProcess(data *int) int {
	*data = *data * 2
	return *data
}

func main() {
	// 同包调用
	x := 10
	r1 := localProcess(&x)
	fmt.Println(r1)

	// 跨包调用
	y := 20
	r2 := internal.ProcessData(&y)
	fmt.Println(r2)
}
