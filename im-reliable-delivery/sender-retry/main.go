// E1：发送端带唯一 msgID 的超时补传
//
// 演示要点：弱网下 ACK 可能丢失，发送端会重发。关键在于重发时 msgID
// 保持不变——重试不是新消息，接收端可以据此去重。这是 at-least-once
// 投递的发送端保障。
//
// 运行：go run .  （Go 1.26，仅用标准库）
package main

import (
	"fmt"
	"time"
)

// Message 是发送端发出的一条消息。MsgID 在多次重试中保持不变。
type Message struct {
	MsgID string
	Body  string
	Try   int // 第几次尝试（仅用于日志，不参与去重）
}

func main() {
	const msgID = "m-0f3a9c" // 发送端生成的全局唯一 ID，重试不变
	body := "在吗？周末爬山"

	// 模拟"网络"：第 1 次尝试的 ACK 被丢弃，第 2 次正常回 ACK。
	// 真实环境里哪次丢是不可预测的，这里固定第 1 次丢只为演示可见的重试。
	timeout := 50 * time.Millisecond
	for attempt := 1; attempt <= 3; attempt++ {
		msg := Message{MsgID: msgID, Body: body, Try: attempt}
		fmt.Printf("[发送端] 第 %d 次发送 msgID=%s body=%q\n", attempt, msg.MsgID, msg.Body)

		ackCh := make(chan bool, 1)
		go func(a int) {
			if a == 1 {
				fmt.Printf("[网络] 第 %d 次 ACK 丢失\n", a)
				return // 不回 ACK，触发发送端超时重发
			}
			time.Sleep(20 * time.Millisecond) // 模拟网络回程
			ackCh <- true
		}(attempt)

		select {
		case <-ackCh:
			fmt.Printf("[发送端] 收到 ACK，投递成功（共尝试 %d 次，msgID 始终=%s）\n", attempt, msgID)
			return
		case <-time.After(timeout):
			fmt.Printf("[发送端] 超时未收到 ACK，准备用同一 msgID 重发\n")
		}
	}
	fmt.Println("[发送端] 达到最大重试次数，转入离线暂存/失败队列")
}
