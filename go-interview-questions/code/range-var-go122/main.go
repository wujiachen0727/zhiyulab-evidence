package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
)

func main() {
	fmt.Printf("go.mod 声明的语言版本：go 1.22\n")
	fmt.Printf("实际工具链 runtime.Version() = %s\n\n", runtime.Version())

	fmt.Println("== 场景 1：闭包捕获循环变量 ==")
	funcs := make([]func(), 0, 3)
	for _, v := range []int{1, 2, 3} {
		funcs = append(funcs, func() { fmt.Print(v, " ") })
	}
	for _, f := range funcs {
		f()
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("== 场景 2：三次迭代里循环变量的地址 ==")
	addrs := make([]string, 0, 3)
	for i := range []int{1, 2, 3} {
		addrs = append(addrs, fmt.Sprintf("%p", &i))
	}
	fmt.Printf("  三次迭代的 &i：%v\n", addrs)
	fmt.Printf("  地址是否完全相同：%v\n", addrs[0] == addrs[1] && addrs[1] == addrs[2])

	fmt.Println()
	fmt.Println("== 场景 3：协程读循环变量（结果已排序，排除调度顺序干扰）==")
	var (
		mu   sync.Mutex
		seen []int
		wg   sync.WaitGroup
	)
	for _, v := range []int{1, 2, 3} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			seen = append(seen, v)
		}()
	}
	wg.Wait()
	sort.Ints(seen)
	fmt.Printf("  三个协程读到的值：%v\n", seen)

	fmt.Println()
	fmt.Println("== 场景 4：defer 里读循环变量 ==")
	printDeferred()
}

func printDeferred() {
	for _, v := range []int{1, 2, 3} {
		defer fmt.Printf("  defer 读到 %d\n", v)
	}
	fmt.Println("  循环结束，下面才轮到 defer 按后进先出执行")
}
