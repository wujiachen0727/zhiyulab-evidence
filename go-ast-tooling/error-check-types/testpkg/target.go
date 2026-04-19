// 测试目标文件：包含多种 error 忽略场景
package testpkg

import (
	"fmt"
	"os"
)

func goodExample() {
	f, err := os.Open("file.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
}

func badExample1() {
	// 场景1：用 _ 显式丢弃 error
	_, _ = fmt.Println("hello")
}

func badExample2() {
	// 场景2：完全丢弃返回值（纯 AST 看不出这里返回了 error）
	os.Remove("temp.txt")
}

func badExample3() {
	// 场景3：多返回值只取第一个
	f, _ := os.Open("data.txt")
	_ = f
}

func noErrorReturn() {
	// 场景4：函数不返回 error（纯 AST 会误报，go/types 不会）
	fmt.Println("no error here")
}
