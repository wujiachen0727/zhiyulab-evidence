package main

import (
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// 演示：用 ListenConfig.Control 在 bind 前设置 SO_REUSEPORT
// 以及用 syscall.RawConn 在连接建立后设置 TCP_KEEPIDLE

func main() {
	// === 第二层：ListenConfig.Control ===
	// 标准库的 net.Listen 不暴露 SO_REUSEPORT
	// 但你可以用 ListenConfig 的 Control 回调插手

	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				// 设置 SO_REUSEPORT：多个进程/goroutine 可以绑定同一个端口
				// 内核自动做负载均衡
				err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					fmt.Printf("SO_REUSEPORT 设置失败: %v\n", err)
				} else {
					fmt.Println("✅ SO_REUSEPORT 已开启")
				}
			})
		},
	}

	ln, err := lc.Listen(nil, "tcp", ":0")
	if err != nil {
		fmt.Printf("listen error: %v\n", err)
		return
	}
	fmt.Printf("监听地址: %s\n\n", ln.Addr())
	ln.Close()

	// === 第二层：RawConn 设置 TCP_KEEPIDLE ===
	// 标准库只提供 SetKeepAlive(bool) 和 SetKeepAlivePeriod(duration)
	// 但 TCP_KEEPIDLE（空闲多久后开始探测）需要 syscall

	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		fmt.Printf("dial error: %v\n", err)
		return
	}
	defer conn.Close()

	tcpConn := conn.(*net.TCPConn)
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		fmt.Printf("SyscallConn error: %v\n", err)
		return
	}

	rawConn.Control(func(fd uintptr) {
		// TCP_KEEPIDLE: 连接空闲 30 秒后开始发送 keepalive 探测
		// 默认值通常是 7200 秒（2小时），对于需要快速检测断连的场景远远不够
		err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPALIVE, 30)
		if err != nil {
			fmt.Printf("TCP_KEEPIDLE 设置失败: %v\n", err)
		} else {
			fmt.Println("✅ TCP_KEEPIDLE 设为 30 秒（默认 7200 秒）")
		}
	})

	fmt.Println("\n这些参数标准库的 SetXxx 方法无法设置，必须走 syscall。")
}
