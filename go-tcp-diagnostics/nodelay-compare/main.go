// TCP_NODELAY 在不同网络条件下的效果对比
// 本地回环 vs 模拟 1ms 延迟 vs 模拟 10ms 延迟
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
)

// 模拟网络延迟的 Listener 包装
type delayConn struct {
	net.Conn
	delay time.Duration
}

func (c *delayConn) Read(b []byte) (int, error) {
	time.Sleep(c.delay)
	return c.Conn.Read(b)
}

func runBenchmark(noDelay bool, msgSize int, msgCount int, readDelay time.Duration) time.Duration {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	var wg sync.WaitGroup
	wg.Add(1)

	// 接收端
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		var reader net.Conn = conn
		if readDelay > 0 {
			reader = &delayConn{Conn: conn, delay: readDelay}
		}

		buf := make([]byte, 65536)
		totalRead := 0
		for totalRead < msgSize*msgCount {
			n, err := reader.Read(buf)
			if err != nil {
				break
			}
			totalRead += n
		}
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
	wg.Wait()
	return time.Since(start)
}

func main() {
	fmt.Println("=== TCP_NODELAY 多网络环境效果对比 ===")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// 小包场景（最能体现 Nagle 算法影响）
	msgSize := 64
	msgCount := 1000

	delays := []struct {
		label string
		delay time.Duration
	}{
		{"本地回环(0ms)", 0},
		{"模拟1ms延迟", 1 * time.Millisecond},
		{"模拟10ms延迟", 10 * time.Millisecond},
	}

	fmt.Printf("场景：小包(%dB) x %d\n\n", msgSize, msgCount)
	fmt.Printf("%-20s %-15s %-15s %-10s\n", "网络环境", "NoDelay=true", "NoDelay=false", "差异")
	fmt.Println("--------------------------------------------------------------")

	for _, d := range delays {
		// 预热
		runBenchmark(true, msgSize, 100, d.delay)
		runBenchmark(false, msgSize, 100, d.delay)

		// 正式测量（取 3 次均值）
		var sumTrue, sumFalse time.Duration
		rounds := 3
		for i := 0; i < rounds; i++ {
			sumTrue += runBenchmark(true, msgSize, msgCount, d.delay)
			sumFalse += runBenchmark(false, msgSize, msgCount, d.delay)
		}
		avgTrue := sumTrue / time.Duration(rounds)
		avgFalse := sumFalse / time.Duration(rounds)

		diff := float64(avgTrue-avgFalse) / float64(avgFalse) * 100

		fmt.Printf("%-20s %-15s %-15s %+.1f%%\n",
			d.label,
			avgTrue.Truncate(time.Microsecond),
			avgFalse.Truncate(time.Microsecond),
			diff)
	}
}
