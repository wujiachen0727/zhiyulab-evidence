// E6: 生产级心跳状态机实现骨架
// 演示 healthy → suspect → dead → reconnect 状态转换
// 包含：Jitter、连续失败判定、context 优雅退出
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// HeartbeatState 心跳状态
type HeartbeatState int

const (
	StateHealthy  HeartbeatState = iota // 正常收到 ACK
	StateSuspect                        // 连续失败 < 阈值
	StateDead                           // 连续失败 >= 阈值
	StateReconnecting                   // 正在重连
)

func (s HeartbeatState) String() string {
	switch s {
	case StateHealthy:
		return "healthy"
	case StateSuspect:
		return "suspect"
	case StateDead:
		return "dead"
	case StateReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Interval          time.Duration // 心跳间隔
	Jitter            time.Duration // 随机偏移范围
	Timeout           time.Duration // 单次心跳超时
	FailureThreshold  int           // 连续失败阈值（推荐 3）
	ReconnectInterval time.Duration // 重连间隔
}

// DefaultConfig 默认配置（基于 E5 实验数据：阈值=3 误杀率 0.02%）
func DefaultConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Interval:          30 * time.Second,
		Jitter:            5 * time.Second,
		Timeout:           5 * time.Second,
		FailureThreshold:  3,
		ReconnectInterval: 10 * time.Second,
	}
}

// HeartbeatManager 心跳管理器
type HeartbeatManager struct {
	config    HeartbeatConfig
	state     HeartbeatState
	failures  int           // 连续失败次数
	lastACK   time.Time     // 最后一次收到 ACK 的时间
	ticker    *time.Ticker  // 可复用 Ticker（不是 time.After）
	cancel    context.CancelFunc
	onStateChange func(from, to HeartbeatState)
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(config HeartbeatConfig) *HeartbeatManager {
	return &HeartbeatManager{
		config:  config,
		state:   StateHealthy,
		failures: 0,
		lastACK: time.Now(),
	}
}

// Run 启动心跳（阻塞，通过 context 取消退出）
func (h *HeartbeatManager) Run(ctx context.Context) error {
	// Jitter: 30s ± 5s
	interval := h.config.Interval + time.Duration(rand.Intn(int(h.config.Jitter*2)))-h.config.Jitter
	if interval < h.config.Jitter {
		interval = h.config.Jitter
	}

	h.ticker = time.NewTicker(interval)
	defer h.ticker.Stop() // 铁律：创建 Ticker 立即 defer Stop

	for {
		select {
		case <-ctx.Done():
			// context 取消 → 优雅退出
			h.setState(StateDead)
			return ctx.Err()

		case <-h.ticker.C:
			// 发送心跳并等待 ACK
			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, h.config.Timeout)
			
			// 模拟发送心跳
			err := h.sendHeartbeat(timeoutCtx)
			timeoutCancel()

			if err != nil {
				// 心跳失败
				h.failures++
				if h.failures >= h.config.FailureThreshold {
					h.setState(StateDead)
					// 触发重连
					h.reconnect(ctx)
				} else {
					h.setState(StateSuspect)
				}
			} else {
				// 心跳成功
				h.failures = 0
				h.lastACK = time.Now()
				h.setState(StateHealthy)
			}
		}
	}
}

// sendHeartbeat 模拟发送心跳并等待 ACK
func (h *HeartbeatManager) sendHeartbeat(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 超时
	case <-time.After(100 * time.Millisecond):
		return nil // 模拟成功收到 ACK
	}
}

// reconnect 重连逻辑
func (h *HeartbeatManager) reconnect(ctx context.Context) {
	h.setState(StateReconnecting)
	
	reconnectTicker := time.NewTicker(h.config.ReconnectInterval)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-reconnectTicker.C:
			// 尝试重连...
			// 成功则重置状态
			h.failures = 0
			h.lastACK = time.Now()
			h.setState(StateHealthy)
			return
		}
	}
}

// setState 状态转换
func (h *HeartbeatManager) setState(newState HeartbeatState) {
	if h.state == newState {
		return
	}
	oldState := h.state
	h.state = newState
	if h.onStateChange != nil {
		h.onStateChange(oldState, newState)
	}
}

func main() {
	// 演示心跳状态机运行
	config := DefaultConfig()
	hm := NewHeartbeatManager(config)
	hm.onStateChange = func(from, to HeartbeatState) {
		fmt.Printf("[%s] 状态转换: %s → %s\n", time.Now().Format("15:04:05"), from, to)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("启动心跳状态机 (间隔=%v, 阈值=%d)\n", config.Interval, config.FailureThreshold)
	fmt.Println("---")

	// 运行 5 秒后自动退出
	if err := hm.Run(ctx); err != nil {
		fmt.Printf("\n心跳管理器退出: %v\n", err)
	}
}
