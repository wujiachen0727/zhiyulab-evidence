package main

import "fmt"

// attempt 执行 f，若发生 panic 则打印 panic 内容
func attempt(label string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%-40s panic: %v\n", label, r)
		}
	}()
	f()
	fmt.Printf("%-40s 正常执行完毕\n", label)
}

func showMapPtr(label string, p *map[string]int) {
	fmt.Printf("%-40s 类型 %T，指向的 map 是 nil: %v\n", label, p, *p == nil)
}

func main() {
	fmt.Println("== new(T)：分配内存、置零值、返回指针 ==")

	pm := new(map[string]int)
	showMapPtr("new(map[string]int)", pm)

	ps := new([]int)
	fmt.Printf("%-40s 类型 %T，指向的 slice 是 nil: %v\n", "new([]int)", ps, *ps == nil)

	fmt.Println()
	fmt.Println("== nil map：读可以，写不行 ==")

	var nilMap map[string]int
	fmt.Printf("%-40s %v\n", "读 nil map 的键", nilMap["missing"])
	fmt.Printf("%-40s %d\n", "nil map 的 len", len(nilMap))

	attempt("向 nil map 写入", func() {
		var m map[string]int
		m["a"] = 1
	})

	attempt("向 *new(map[string]int) 写入", func() {
		m := new(map[string]int)
		(*m)["a"] = 1
	})

	fmt.Println()
	fmt.Println("== nil slice：可以直接 append，这是有意设计 ==")

	var nilSlice []int
	fmt.Printf("%-40s len=%d cap=%d 是 nil: %v\n", "var s []int", len(nilSlice), cap(nilSlice), nilSlice == nil)
	nilSlice = append(nilSlice, 1, 2)
	fmt.Printf("%-40s len=%d cap=%d 值=%v\n", "append 之后", len(nilSlice), cap(nilSlice), nilSlice)

	fmt.Println()
	fmt.Println("== make(T)：构造可用的引用类型 ==")

	m := make(map[string]int)
	m["a"] = 1
	fmt.Printf("%-40s %v\n", "make(map[string]int) 写入后", m)

	s := make([]int, 0, 4)
	fmt.Printf("%-40s len=%d cap=%d 是 nil: %v\n", "make([]int, 0, 4)", len(s), cap(s), s == nil)
}
