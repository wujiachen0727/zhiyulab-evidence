// channel send 接收方退出导致的永久阻塞（无缓冲 channel 版本）
// 运行：go run chan_send_block_unbuffered.go
//
// 用无缓冲 channel 清晰展示：接收方退出后，发送方永久阻塞
// goroutine dump 特征：goroutine 卡在 "chan send"

package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	// 无缓冲 channel —— 发送必须等接收
	ch := make(chan string)

	// 启动一个接收方，接收 2 条消息后退出了
	go func() {
		for i := 0; i < 2; i++ {
			msg := <-ch
			fmt.Printf("[Receiver] 收到: %s\n", msg)
		}
		fmt.Println("[Receiver] 接收完毕，退出！")
		// 接收方退出了，但发送方不知道
	}()

	time.Sleep(50 * time.Millisecond)

	// 发送方发送 3 条消息
	sendDone := make(chan struct{})
	go func() {
		defer close(sendDone)
		msgs := []string{"msg1", "msg2", "msg3"}
		for _, msg := range msgs {
			fmt.Printf("[Sender] 发送: %s\n", msg)
			ch <- msg // 发送 msg3 时，已经没有接收方了 —— 永久阻塞！
		}
		fmt.Println("[Sender] 全部发送完成")
	}()

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== goroutine dump 特征信号 ===")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	dump := string(buf[:n])

	if contains(dump, "chan send") {
		fmt.Println("✅ 发现 chan send 信号（channel send 阻塞的特征）")
	}

	fmt.Println("\n--- 关键 dump 片段 ---")
	lines := splitLines(dump)
	printContext(lines, "chan send", 2)

	fmt.Println("\n⚠️ 注意：Sender 已永久阻塞在 ch <- \"msg3\"")
	fmt.Println("⚠️ 接收方已退出，但发送方不知道——channel 不提供'对方还在吗'的查询")

	// 为了让程序能退出，额外启动一个接收方读取剩余消息
	go func() {
		for msg := range ch {
			fmt.Printf("[ExtraReceiver] 取走剩余消息: %s\n", msg)
		}
	}()

	<-sendDone
	fmt.Println("✅ 程序退出")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func printContext(lines []string, keyword string, contextLines int) {
	for i, line := range lines {
		if contains(line, keyword) {
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			end := i + contextLines + 1
			if end > len(lines) {
				end = len(lines)
			}
			for j := start; j < end; j++ {
				fmt.Println(lines[j])
			}
			fmt.Println("---")
		}
	}
}
