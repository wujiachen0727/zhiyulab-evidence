package main

import (
	"fmt"
	"time"
)

// attempt 先打印用例标签，再执行 f，捕获 panic 后单独一行显示
func attempt(label string, f func()) {
	fmt.Printf("  %s\n", label)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    -> panic: %v\n", r)
		}
	}()
	f()
}

func main() {
	fmt.Println("== 1. 关闭之后继续接收 ==")
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)
	for i := 1; i <= 4; i++ {
		v, ok := <-ch
		fmt.Printf("    第 %d 次接收：值=%d ok=%v\n", i, v, ok)
	}

	fmt.Println()
	fmt.Println("== 2. 关闭之后继续发送 ==")
	attempt("向已关闭的 channel 发送", func() {
		c := make(chan int)
		close(c)
		c <- 1
	})

	fmt.Println()
	fmt.Println("== 3. 重复关闭 ==")
	attempt("同一个 channel 关闭两次", func() {
		c := make(chan int)
		close(c)
		close(c)
	})

	fmt.Println()
	fmt.Println("== 4. nil channel 上的收发都永久阻塞 ==")
	attempt("在 nil channel 上接收（配 150ms 超时）", func() {
		var c chan int
		select {
		case <-c:
			fmt.Println("    -> 收到了值")
		case <-time.After(150 * time.Millisecond):
			fmt.Println("    -> 150ms 内没有任何进展，接收永久阻塞")
		}
	})
	attempt("在 nil channel 上发送（配 150ms 超时）", func() {
		var c chan int
		select {
		case c <- 1:
			fmt.Println("    -> 发出去了")
		case <-time.After(150 * time.Millisecond):
			fmt.Println("    -> 150ms 内没有任何进展，发送永久阻塞")
		}
	})

	fmt.Println()
	fmt.Println("== 5. 接收方关闭 channel，发送方会崩 ==")
	attempt("接收方先 close，发送方随后发送", func() {
		c := make(chan int)
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("    -> 发送方 panic: %v\n", r)
				}
			}()
			time.Sleep(50 * time.Millisecond)
			c <- 1
		}()
		close(c)
		<-done
	})
}
