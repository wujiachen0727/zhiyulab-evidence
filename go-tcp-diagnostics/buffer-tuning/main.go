// SO_RCVBUF / SO_SNDBUF 调优对比实验
// 测量不同缓冲区大小对吞吐量和延迟的影响
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"fmt"
	"net"
	"runtime"
	"syscall"
	"time"
)

func setBufferSize(conn net.Conn, rcvBuf, sndBuf int) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("not a TCP connection")
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return err
	}

	return rawConn.Control(func(fd uintptr) {
		if rcvBuf > 0 {
			syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, rcvBuf)
		}
		if sndBuf > 0 {
			syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, sndBuf)
		}
	})
}

func getBufferSize(conn net.Conn) (rcvBuf, sndBuf int) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, 0
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return 0, 0
	}

	rawConn.Control(func(fd uintptr) {
		rcvBuf, _ = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		sndBuf, _ = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF)
	})
	return
}

func runThroughput(totalBytes int, chunkSize int, rcvBuf, sndBuf int) (throughput float64, duration time.Duration) {
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
			close(done)
			return
		}
		defer conn.Close()

		if rcvBuf > 0 || sndBuf > 0 {
			setBufferSize(conn, rcvBuf, sndBuf)
		}

		buf := make([]byte, 65536)
		totalRead := 0
		for totalRead < totalBytes {
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

	if rcvBuf > 0 || sndBuf > 0 {
		setBufferSize(conn, rcvBuf, sndBuf)
	}

	data := make([]byte, chunkSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	totalSent := 0
	start := time.Now()
	for totalSent < totalBytes {
		n, err := conn.Write(data)
		if err != nil {
			break
		}
		totalSent += n
	}
	conn.Close()
	<-done
	duration = time.Since(start)
	throughput = float64(totalBytes) / duration.Seconds() / 1024 / 1024 // MB/s
	return
}

func main() {
	fmt.Println("=== SO_RCVBUF/SO_SNDBUF 调优对比实验 ===")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// 先看默认缓冲区大小
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	conn, _ := net.Dial("tcp", ln.Addr().String())
	srvConn, _ := ln.Accept()
	rcv, snd := getBufferSize(conn)
	srvRcv, srvSnd := getBufferSize(srvConn)
	fmt.Printf("系统默认缓冲区 — 客户端: SO_RCVBUF=%d, SO_SNDBUF=%d\n", rcv, snd)
	fmt.Printf("                 服务端: SO_RCVBUF=%d, SO_SNDBUF=%d\n", srvRcv, srvSnd)
	conn.Close()
	srvConn.Close()
	ln.Close()
	fmt.Println()

	totalBytes := 100 * 1024 * 1024 // 100MB
	chunkSize := 4096               // 4KB per write

	bufferConfigs := []struct {
		label  string
		rcvBuf int
		sndBuf int
	}{
		{"系统默认", 0, 0},
		{"8KB", 8 * 1024, 8 * 1024},
		{"64KB", 64 * 1024, 64 * 1024},
		{"256KB", 256 * 1024, 256 * 1024},
		{"1MB", 1024 * 1024, 1024 * 1024},
	}

	fmt.Printf("传输 %dMB 数据（每次写 %dB）\n\n", totalBytes/1024/1024, chunkSize)
	fmt.Printf("%-15s %-15s %-15s\n", "缓冲区大小", "吞吐量(MB/s)", "耗时")
	fmt.Println("----------------------------------------------")

	for _, cfg := range bufferConfigs {
		// 取 3 次均值
		var totalTP float64
		var totalDur time.Duration
		rounds := 3
		for i := 0; i < rounds; i++ {
			tp, dur := runThroughput(totalBytes, chunkSize, cfg.rcvBuf, cfg.sndBuf)
			totalTP += tp
			totalDur += dur
		}
		avgTP := totalTP / float64(rounds)
		avgDur := totalDur / time.Duration(rounds)

		fmt.Printf("%-15s %-15.1f %-15s\n",
			cfg.label, avgTP, avgDur.Truncate(time.Millisecond))
	}
}
