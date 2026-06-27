// channel send 接收方退出导致的永久阻塞
// 运行：go run chan_send_block.go
//
// 典型场景：worker pool 中，接收方因错误退出后，发送方继续发送
// - 创建一个有缓冲的 channel
// - 多个 worker goroutine 处理任务
// - 某个 worker 出错退出 → channel 的接收方减少
// - 发送方还在继续发送 → 积累到 channel 满 → 永久阻塞
//
// goroutine dump 特征：大批 goroutine 卡在 "chan send"

package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	const numWorkers = 2
	const numTasks = 10

	tasks := make(chan int, 3) // 小缓冲更容易触发阻塞
	done := make(chan struct{})

	// 启动 worker
	for i := 0; i < numWorkers; i++ {
		id := i
		go func() {
			for task := range tasks {
				if task == 4 && id == 0 { // 第 0 号 worker 在第 5 个任务时退出
					fmt.Printf("[Worker%d] 遇到错误，退出！\n", id)
					return // 直接退出
				}
				fmt.Printf("[Worker%d] 处理任务 %d\n", id, task)
				time.Sleep(30 * time.Millisecond)
			}
		}()
	}

	// 让 worker 先就绪
	time.Sleep(50 * time.Millisecond)

	// 发送任务（在单独的 goroutine 中）
	go func() {
		for i := 0; i < numTasks; i++ {
			fmt.Printf("[Sender] 发送任务 %d\n", i)
			tasks <- i
		}
		close(tasks)
		fmt.Println("[Sender] 所有任务发送完成")
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== goroutine dump 特征信号 ===")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	dump := string(buf[:n])

	if contains(dump, "chan send") {
		fmt.Println("✅ 发现 chan send 信号（channel send 阻塞的特征）")
	}
	if contains(dump, "tasks<-") {
		fmt.Println("✅ 发现发送方卡在 tasks<- 语句")
	}

	fmt.Println("\n--- 关键 dump 片段 ---")
	lines := splitLines(dump)
	for _, line := range lines {
		if contains(line, "chan send") || contains(line, "chan receive") || contains(line, "tasks") {
			fmt.Println(line)
		}
	}

	<-done
	time.Sleep(50 * time.Millisecond)
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
