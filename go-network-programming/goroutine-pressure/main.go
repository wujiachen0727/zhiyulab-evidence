// goroutine-per-conn 内存压力测试
// 模拟不同数量的网络连接，测量每个连接的内存开销
// [实测 Go 1.26.2 darwin/arm64]
package main

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

func measureMemory() uint64 {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func main() {
	fmt.Println("=== goroutine-per-conn 内存压力测试 ===")
	fmt.Println("Go 版本:", runtime.Version())
	fmt.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
	fmt.Println()

	// 启动一个 TCP 监听器
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := ln.Addr().String()
	fmt.Println("监听地址:", addr)

	// 服务端：接受连接后挂起 goroutine（模拟 goroutine-per-conn）
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				buf := make([]byte, 4096) // 典型的读缓冲区
				for {
					_, err := c.Read(buf)
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	// 测试不同连接数下的内存占用
	testCases := []int{100, 1000, 5000, 10000}

	fmt.Printf("%-12s %-15s %-15s %-15s %-15s\n",
		"连接数", "总内存(MB)", "增量内存(MB)", "每连接(KB)", "Goroutines")
	fmt.Println("--------------------------------------------------------------------")

	baseMem := measureMemory()
	baseGoroutines := runtime.NumGoroutine()

	var allConns []net.Conn

	for _, count := range testCases {
		// 建立连接到目标数量
		for len(allConns) < count {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("连接失败 at %d: %v\n", len(allConns), err)
				break
			}
			allConns = append(allConns, conn)
		}

		time.Sleep(200 * time.Millisecond) // 等待 goroutine 启动

		currentMem := measureMemory()
		goroutines := runtime.NumGoroutine()
		deltaMem := currentMem - baseMem
		perConn := float64(deltaMem) / float64(len(allConns)) / 1024

		fmt.Printf("%-12d %-15.2f %-15.2f %-15.2f %-15d\n",
			len(allConns),
			float64(currentMem)/1024/1024,
			float64(deltaMem)/1024/1024,
			perConn,
			goroutines)
	}

	fmt.Println()
	fmt.Println("基线信息:")
	fmt.Printf("  基线内存: %.2f MB\n", float64(baseMem)/1024/1024)
	fmt.Printf("  基线 Goroutines: %d\n", baseGoroutines)

	// 清理
	for _, c := range allConns {
		c.Close()
	}
	ln.Close()

	time.Sleep(500 * time.Millisecond)
	finalMem := measureMemory()
	fmt.Printf("  清理后内存: %.2f MB\n", float64(finalMem)/1024/1024)
	fmt.Printf("  清理后 Goroutines: %d\n", runtime.NumGoroutine())
}
