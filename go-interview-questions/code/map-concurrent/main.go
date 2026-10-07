package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("== 1. 单协程写 nil map：是 panic，recover 接得住 ==")

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  结果：panic 被 recover 接住 -> %v\n", r)
			}
		}()
		var m map[string]int
		m["a"] = 1
	}()

	fmt.Println()
	fmt.Println("== 2. 多协程并发写同一个 map：fatal error ==")
	fmt.Println("  8 个协程各写 10 万次，每个协程内部都装了 recover")

	m := make(map[int]int)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  recover 接住了：%v\n", r)
				}
			}()
			for j := 0; j < 100000; j++ {
				m[j] = id
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("  走到这一行说明没有发生 fatal error")
}
