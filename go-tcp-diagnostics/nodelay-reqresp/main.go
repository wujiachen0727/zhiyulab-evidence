// TCP_NODELAY 在请求-响应模式下的效果对比
// Nagle 算法对延迟的影响在"发一个小包、等回复"的模式下最明显
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"time"
)

func runReqResp(noDelay bool, msgSize int, rounds int) (avgLatency time.Duration, totalTime time.Duration) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	// 服务端：收到请求后立即回复
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetNoDelay(noDelay)
		}

		buf := make([]byte, msgSize+4)
		for {
			// 读取请求（4字节长度头 + payload）
			_, err := conn.Read(buf)
			if err != nil {
				return
			}
			// 发送响应（固定 4 字节 "OK\n\x00"）
			conn.Write([]byte("OK\n\x00"))
		}
	}()

	// 客户端：发请求、等回复、计时
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(noDelay)
	}

	msg := make([]byte, 4+msgSize)
	binary.BigEndian.PutUint32(msg[:4], uint32(msgSize))
	for i := 4; i < len(msg); i++ {
		msg[i] = byte(i % 256)
	}
	respBuf := make([]byte, 64)

	// 预热
	for i := 0; i < 50; i++ {
		conn.Write(msg)
		conn.Read(respBuf)
	}

	var totalLatency time.Duration
	start := time.Now()
	for i := 0; i < rounds; i++ {
		reqStart := time.Now()
		conn.Write(msg)
		conn.Read(respBuf)
		totalLatency += time.Since(reqStart)
	}
	totalTime = time.Since(start)
	avgLatency = totalLatency / time.Duration(rounds)
	return
}

func main() {
	fmt.Println("=== TCP_NODELAY 请求-响应延迟对比 ===")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	rounds := 5000

	scenarios := []struct {
		label   string
		msgSize int
	}{
		{"小请求(32B)", 32},
		{"小请求(64B)", 64},
		{"中请求(512B)", 512},
		{"大请求(4KB)", 4096},
	}

	fmt.Printf("每场景 %d 次请求-响应往返\n\n", rounds)
	fmt.Printf("%-18s %-18s %-18s %-12s\n", "场景", "NoDelay=true", "NoDelay=false", "差异")
	fmt.Println("------------------------------------------------------------------")

	for _, s := range scenarios {
		latTrue, _ := runReqResp(true, s.msgSize, rounds)
		latFalse, _ := runReqResp(false, s.msgSize, rounds)

		diff := float64(latFalse-latTrue) / float64(latFalse) * 100

		fmt.Printf("%-18s %-18s %-18s %.1f%%\n",
			s.label,
			latTrue.Truncate(time.Microsecond),
			latFalse.Truncate(time.Microsecond),
			diff)
	}

	fmt.Println("\n说明：NoDelay=true 在请求-响应模式下通常更快（减少 Nagle 等待），")
	fmt.Println("但在单向批量发送模式下可能更慢（增加系统调用次数）。")
	fmt.Println("结论：TCP_NODELAY 不是万金油——适用场景决定效果。")
}
