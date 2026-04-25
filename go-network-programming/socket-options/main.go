// TCP_NODELAY 实际效果对比实验
// 测量启用/禁用 Nagle 算法对小包发送延迟的影响
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

func runTest(noDelay bool, msgSize int, msgCount int) time.Duration {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	done := make(chan struct{})

	// 接收端
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 65536)
		totalRead := 0
		for totalRead < msgSize*msgCount {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			totalRead += n
		}
		close(done)
	}()

	// 发送端
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(noDelay)
	}

	msg := make([]byte, msgSize)
	for i := range msg {
		msg[i] = byte(i % 256)
	}

	start := time.Now()
	for i := 0; i < msgCount; i++ {
		conn.Write(msg)
	}
	conn.Close()
	<-done
	return time.Since(start)
}

func main() {
	fmt.Println("=== TCP_NODELAY 效果对比实验 ===")
	fmt.Println("Go 版本:", runtime.Version())
	fmt.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
	fmt.Println()

	testCases := []struct {
		msgSize  int
		msgCount int
		label    string
	}{
		{64, 10000, "小包(64B) x 10000"},
		{512, 10000, "中包(512B) x 10000"},
		{4096, 5000, "大包(4KB) x 5000"},
	}

	fmt.Printf("%-25s %-15s %-15s %-10s\n", "场景", "NoDelay=true", "NoDelay=false", "差异")
	fmt.Println("------------------------------------------------------------------")

	for _, tc := range testCases {
		// 预热
		runTest(true, tc.msgSize, 100)
		runTest(false, tc.msgSize, 100)

		// 正式测量（取 3 次均值）
		var sumTrue, sumFalse time.Duration
		rounds := 3
		for i := 0; i < rounds; i++ {
			sumTrue += runTest(true, tc.msgSize, tc.msgCount)
			sumFalse += runTest(false, tc.msgSize, tc.msgCount)
		}
		avgTrue := sumTrue / time.Duration(rounds)
		avgFalse := sumFalse / time.Duration(rounds)

		diff := float64(avgFalse-avgTrue) / float64(avgFalse) * 100

		fmt.Printf("%-25s %-15s %-15s %.1f%%\n",
			tc.label,
			avgTrue.Truncate(time.Microsecond),
			avgFalse.Truncate(time.Microsecond),
			diff)
	}
}
