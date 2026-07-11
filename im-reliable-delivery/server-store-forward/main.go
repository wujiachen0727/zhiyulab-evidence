// E2：服务端 store-and-forward（离线暂存 + 上线拉历史）
//
// 演示要点：服务端收到消息先落库，再推在线设备；离线设备在库里暂存，
// 等它上线时拉取历史消息。这是 IM 在"设备离线"场景下的第二道防线。
//
// 运行：go run .  （Go 1.26，仅用标准库）
//
// 简化说明：真实系统按"每设备已读游标"增量拉取，本 demo 为演示清晰
// 让离线设备上线时拉取全量历史。生产环境不会把已读消息再推一遍。
package main

import (
	"fmt"
	"sync"
)

// Device 表示一个登录设备。
type Device struct {
	ID     string
	Online bool
	Inbox  []string
}

// Server 持有消息库和各设备状态。
type Server struct {
	mu      sync.Mutex
	store   []string         // 落库的消息（模拟 DB）
	devices map[string]*Device
}

// Receive 服务端收到一条消息：先落库，再推在线设备，离线设备暂存。
func (s *Server) Receive(msg string) {
	s.mu.Lock()
	s.store = append(s.store, msg) // 1. 落库
	seq := len(s.store)
	s.mu.Unlock()
	fmt.Printf("[服务端] 消息落库 seq=%d: %q\n", seq, msg)

	// 2. 推在线设备；离线设备等待上线拉取
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.devices {
		if d.Online {
			d.Inbox = append(d.Inbox, msg)
			fmt.Printf("[服务端] 推送给在线设备 %s\n", d.ID)
		} else {
			fmt.Printf("[服务端] 设备 %s 离线，暂存，等待上线拉取\n", d.ID)
		}
	}
}

// OnConnect 设备上线：从库里拉取历史未投递消息。
func (s *Server) OnConnect(devID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.devices[devID]
	d.Online = true
	fmt.Printf("[服务端] 设备 %s 上线，拉取历史消息 %d 条\n", devID, len(s.store))
	d.Inbox = append(d.Inbox, s.store...) // 真实场景按已读游标增量拉取
}

func main() {
	s := &Server{devices: map[string]*Device{
		"phone":  {ID: "phone", Online: true},
		"pc":     {ID: "pc", Online: false},
		"tablet": {ID: "tablet", Online: false},
	}}

	s.Receive("晚上八点组局火锅")

	fmt.Println("--- PC 上线 ---")
	s.OnConnect("pc")
	fmt.Println("--- tablet 上线 ---")
	s.OnConnect("tablet")

	fmt.Println("\n各设备收件箱：")
	for _, d := range s.devices {
		fmt.Printf("  %s: %v\n", d.ID, d.Inbox)
	}
}
