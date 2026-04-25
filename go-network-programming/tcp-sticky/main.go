// TCP "粘包"复现实验
// 发送端快速发送多条消息，接收端展示粘包现象
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

func main() {
	fmt.Println("=== TCP 粘包复现实验 ===")
	fmt.Println("Go 版本:", runtime.Version())
	fmt.Println()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := ln.Addr().String()

	done := make(chan struct{})

	// 接收端：用固定 buffer 读取，展示实际收到的内容
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		readCount := 0
		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			readCount++
			fmt.Printf("第 %d 次 Read: 收到 %d 字节 → %q\n", readCount, n, string(buf[:n]))
		}
		fmt.Printf("\n总计 Read 调用次数: %d\n", readCount)
		close(done)
	}()

	// 发送端：快速连续发送 5 条消息
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	messages := []string{
		"msg-1:hello",
		"msg-2:world",
		"msg-3:foo",
		"msg-4:bar",
		"msg-5:baz",
	}

	fmt.Printf("发送端连续发送 %d 条消息:\n", len(messages))
	for i, msg := range messages {
		n, err := conn.Write([]byte(msg))
		if err != nil {
			panic(err)
		}
		fmt.Printf("  发送第 %d 条: %d 字节 → %q\n", i+1, n, msg)
	}

	fmt.Println()
	fmt.Println("接收端实际收到:")
	conn.Close()
	time.Sleep(500 * time.Millisecond)
	<-done

	fmt.Println()
	fmt.Println("结论: 发送了 5 条消息，但接收端的 Read 次数通常少于 5 次。")
	fmt.Println("这就是所谓的'粘包'——TCP 是字节流，不保留消息边界。")
}
