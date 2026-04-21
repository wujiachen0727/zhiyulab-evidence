package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 最小可靠 UDP 协议演示
// 只实现序列号 + ACK + 超时重传，不做拥塞控制
// 目的：展示"按需选配协议属性"的思路

// 协议格式（极简）：
// [0:1]  类型: 0=DATA, 1=ACK
// [1:5]  序列号 (uint32, big endian)
// [5:]   载荷

const (
	typeData = 0
	typeACK  = 1
)

type packet struct {
	typ    byte
	seqNum uint32
	data   []byte
}

func encodePacket(p packet) []byte {
	buf := make([]byte, 5+len(p.data))
	buf[0] = p.typ
	binary.BigEndian.PutUint32(buf[1:5], p.seqNum)
	copy(buf[5:], p.data)
	return buf
}

func decodePacket(buf []byte, n int) packet {
	p := packet{
		typ:    buf[0],
		seqNum: binary.BigEndian.Uint32(buf[1:5]),
	}
	if n > 5 {
		p.data = make([]byte, n-5)
		copy(p.data, buf[5:n])
	}
	return p
}

// 可靠发送器：带超时重传的 stop-and-wait
func reliableSend(conn *net.UDPConn, addr *net.UDPAddr, messages []string) {
	for i, msg := range messages {
		seq := uint32(i)
		pkt := encodePacket(packet{typ: typeData, seqNum: seq, data: []byte(msg)})

		retries := 0
		maxRetries := 3
		timeout := 200 * time.Millisecond

		for retries < maxRetries {
			_, err := conn.WriteToUDP(pkt, addr)
			if err != nil {
				fmt.Printf("[发送] seq=%d 发送失败: %v\n", seq, err)
				return
			}

			// 等待 ACK
			conn.SetReadDeadline(time.Now().Add(timeout))
			buf := make([]byte, 1024)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				retries++
				fmt.Printf("[发送] seq=%d 超时，第 %d 次重传\n", seq, retries)
				timeout *= 2 // 指数退避
				continue
			}

			ack := decodePacket(buf, n)
			if ack.typ == typeACK && ack.seqNum == seq {
				fmt.Printf("[发送] seq=%d ACK 收到 ✅\n", seq)
				break
			}
		}

		if retries >= maxRetries {
			fmt.Printf("[发送] seq=%d 达到最大重试次数，放弃\n", seq)
		}
	}
}

// 接收器
func reliableReceive(conn *net.UDPConn, expected int) []string {
	var received []string
	seen := make(map[uint32]bool)

	for len(received) < expected {
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("[接收] 读取超时\n")
			break
		}

		pkt := decodePacket(buf, n)
		if pkt.typ == typeData {
			// 发送 ACK
			ack := encodePacket(packet{typ: typeACK, seqNum: pkt.seqNum})
			conn.WriteToUDP(ack, addr)

			if !seen[pkt.seqNum] {
				seen[pkt.seqNum] = true
				received = append(received, string(pkt.data))
				fmt.Printf("[接收] seq=%d data=%q ACK 已发送 ✅\n", pkt.seqNum, string(pkt.data))
			} else {
				fmt.Printf("[接收] seq=%d 重复包，已忽略（ACK 已补发）\n", pkt.seqNum)
			}
		}
	}
	return received
}

func main() {
	fmt.Println("=== 最小可靠 UDP 协议演示 ===")
	fmt.Println("协议属性：✅ 序列号  ✅ ACK  ✅ 超时重传  ❌ 拥塞控制  ❌ 流量控制")
	fmt.Println()

	// 启动接收端
	recvAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	recvConn, _ := net.ListenUDP("udp", recvAddr)
	defer recvConn.Close()
	actualAddr := recvConn.LocalAddr().(*net.UDPAddr)

	messages := []string{"hello", "world", "Go网络编程"}

	var wg sync.WaitGroup
	var received []string

	wg.Add(1)
	go func() {
		defer wg.Done()
		received = reliableReceive(recvConn, len(messages))
	}()

	// 启动发送端
	time.Sleep(50 * time.Millisecond)
	sendAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	sendConn, _ := net.ListenUDP("udp", sendAddr)
	defer sendConn.Close()

	reliableSend(sendConn, actualAddr, messages)

	wg.Wait()

	fmt.Printf("\n发送 %d 条，收到 %d 条\n", len(messages), len(received))
	fmt.Println("\n关键点：这个协议只有 ~100 行代码，因为我们只选了需要的属性。")
	fmt.Println("如果你还需要拥塞控制，那就不该自己写——直接用 QUIC。")
}
