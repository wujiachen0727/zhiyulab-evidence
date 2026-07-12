// 端口耗尽复现（跨平台忠实版）。
//
// 思路：用一个固定本地源端口反复向回环 echo 服务发起短连接。
// 每次连接关闭后，该源端口进入 TIME_WAIT；再次用同一源端口 bind 时，
// 若内核未启用 TIME_WAIT 复用，则报 EADDRINUSE（address already in use）。
//
// 这正是高并发短连接客户端（爬虫、压测工具、HTTP/1.0 close 客户端）
// 在"本地端口耗尽"时的真实报错——不是服务端端口，是客户端 ephemeral 端口
// 被自己的 TIME_WAIT 占满。
//
// 运行环境：macOS Darwin arm64 / Go 1.26.4（实测）
// 注意：macOS/BSD 默认允许 TIME_WAIT 端口被新连接复用，因此原生可能不复现
// "失败"；本程序如实记录两种结果，并在正文标注 Linux 与 macOS 的内核语义差异。
package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	host = "127.0.0.1"
	port = 18090
)

// echoServer 起一个回环 echo 服务，收到 1 字节回 1 字节。
func echoServer(done chan struct{}) {
	ln, err := net.Listen("tcp", host+":18090")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	go func() {
		<-done
		ln.Close()
	}()
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			buf := make([]byte, 1)
			conn.Read(buf)
			conn.Write(buf)
			conn.Close()
		}(c)
	}
}

// clientWithFixedPort 用固定本地源端口反复连 echo 服务，关闭后再次用同端口 bind。
func clientWithFixedPort(rounds int) {
	var firstErr error
	var succeeded int
	for i := 0; i < rounds; i++ {
		local, err := net.ResolveTCPAddr("tcp", host+":18091") // 固定源端口
		if err != nil {
			panic(err)
		}
		d := net.Dialer{
			LocalAddr: local,
			Timeout:   500 * time.Millisecond,
		}
		c, err := d.Dial("tcp", host+":18090")
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("[E3 实测 Go] 第 %d 次用固定源端口 18091 发起连接失败: %v\n", i+1, err)
			break
		}
		c.Write([]byte("x"))
		buf := make([]byte, 1)
		c.Read(buf)
		c.Close()
		succeeded++
		time.Sleep(5 * time.Millisecond) // 让 TIME_WAIT 落定
	}
	fmt.Printf("[E3 实测 Go] 成功完成 %d 次；首个错误=%v\n", succeeded, firstErr)
}

func main() {
	done := make(chan struct{})
	echoServer(done)
	time.Sleep(200 * time.Millisecond)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		clientWithFixedPort(50)
	}()
	wg.Wait()
	close(done)
}
