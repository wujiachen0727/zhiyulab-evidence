// E3：接收端多端幂等去重
//
// 演示要点：同一 msgID 推送到一个用户的 N 台设备，每台设备各自去重、
// 各消费一次（消息要出现在我所有设备上，但不能在我某台设备上出现两次）。
// at-least-once 下服务端可能因弱网/重连重推历史消息，去重集保证不重复消费。
//
// 与"通用 MQ 单消费者幂等"的区别：这里强调 IM 的"多端"维度（同一用户多台
// 设备各自有消费状态）和"历史消息"维度（重连重推），而非单消费者 Redis 去重。
//
// 运行：go run .  （Go 1.26，仅用标准库）
//
// 生产环境：seen 集合用 Redis SETNX 或 DB 唯一约束 + 已读游标实现，
// 本 demo 用进程内 map 模拟，保证可独立运行、可复现。
package main

import (
	"fmt"
	"sync"
)

// DeviceInbox 一台设备的收件去重状态。
type DeviceInbox struct {
	ID       string
	seen     map[string]bool // 已消费 msgID 集合（生产用 Redis SETNX / DB 唯一约束）
	consumed int
	mu       sync.Mutex
}

// Deliver 服务端推来一条消息。已见过则丢弃，否则消费一次。
func (d *DeviceInbox) Deliver(msgID, body string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seen[msgID] {
		fmt.Printf("[%s] msgID=%s 已消费过，丢弃（去重）\n", d.ID, msgID)
		return
	}
	d.seen[msgID] = true
	d.consumed++
	fmt.Printf("[%s] 消费 msgID=%s body=%q（本设备第 %d 条）\n", d.ID, msgID, body, d.consumed)
}

func main() {
	phone := &DeviceInbox{ID: "phone", seen: map[string]bool{}}
	pc := &DeviceInbox{ID: "pc", seen: map[string]bool{}}
	tablet := &DeviceInbox{ID: "tablet", seen: map[string]bool{}}

	const msgID = "m-0f3a9c"
	body := "晚上八点组局火锅"

	fmt.Println("=== 场景1：同一消息推给一个用户的三台设备 ===")
	for _, d := range []*DeviceInbox{phone, pc, tablet} {
		d.Deliver(msgID, body) // 三台设备各消费一次
	}

	fmt.Println("\n=== 场景2：at-least-once 下服务端重推同一条历史消息 ===")
	phone.Deliver(msgID, body) // phone 已见过 → 去重
	pc.Deliver(msgID, body)    // pc 已见过 → 去重

	fmt.Printf("\n各设备最终消费条数：phone=%d pc=%d tablet=%d（均为 1，无重复）\n",
		phone.consumed, pc.consumed, tablet.consumed)
}
