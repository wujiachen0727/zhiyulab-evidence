package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// State 限流器有限状态机状态
type State int

const (
	// Normal 正常状态，流量低于阈值
	Normal State = iota
	// SoftLimit 软限流，流量超过正常阈值但未超上限
	SoftLimit
	// HardLimit 硬限流，流量超过每秒上限
	HardLimit
	// Cooldown 冷却期，硬限流结束后短暂恢复窗口
	Cooldown
)

func (s State) String() string {
	switch s {
	case Normal:
		return "NORMAL"
	case SoftLimit:
		return "SOFT_LIMIT"
	case HardLimit:
		return "HARD_LIMIT"
	case Cooldown:
		return "COOLDOWN"
	default:
		return "UNKNOWN"
	}
}

const (
	normalThreshold   = 80               // 每秒超过该值进入 SOFT_LIMIT
	hardThreshold     = 100              // 每秒超过该值进入 HARD_LIMIT，即请求上限
	hardLimitDuration = 10 * time.Second // HARD_LIMIT 持续时长，到期进入 COOLDOWN
	cooldownDuration  = 5 * time.Second  // COOLDOWN 持续时长，到期回到 NORMAL
	windowSize        = time.Second      // 速率统计窗口，1 秒
)

// RateLimiter 带有限状态机的 HTTP 限流器
type RateLimiter struct {
	mu          sync.Mutex // 保护以下字段并发访问
	state       State      // 当前状态
	stateSince  time.Time  // 进入当前状态的时间
	count       int        // 当前 1 秒窗口内放行的请求数
	windowStart time.Time  // 当前统计窗口的起始时间
}

// NewRateLimiter 创建初始状态为 NORMAL 的限流器
func NewRateLimiter() *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		state:       Normal,
		stateSince:  now,
		windowStart: now,
	}
}

// advance 推进统计窗口，窗口跨秒时重置请求计数
func (rl *RateLimiter) advance(now time.Time) {
	if now.Sub(rl.windowStart) >= windowSize {
		rl.windowStart = now
		rl.count = 0
	}
}

// transition 按迁移规则执行状态转移
func (rl *RateLimiter) transition(now time.Time) {
	switch rl.state {
	case Normal:
		switch {
		case rl.count > hardThreshold:
			rl.enter(HardLimit, now)
		case rl.count > normalThreshold:
			rl.enter(SoftLimit, now)
		}
	case SoftLimit:
		switch {
		case rl.count > hardThreshold:
			rl.enter(HardLimit, now)
		case rl.count <= normalThreshold:
			rl.enter(Normal, now)
		}
	case HardLimit:
		if now.Sub(rl.stateSince) >= hardLimitDuration {
			rl.enter(Cooldown, now)
		}
	case Cooldown:
		if now.Sub(rl.stateSince) >= cooldownDuration {
			rl.enter(Normal, now)
		}
	}
}

// enter 切换到目标状态并记录切换时间
func (rl *RateLimiter) enter(s State, now time.Time) {
	rl.state = s
	rl.stateSince = now
}

// Allow 记录一次请求，返回是否放行和当前状态；HARD_LIMIT 下拒绝请求
func (rl *RateLimiter) Allow() (allowed bool, state State) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.advance(now)
	rl.count++
	rl.transition(now)

	state = rl.state
	allowed = state != HardLimit
	return
}

func main() {
	rl := NewRateLimiter()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		allowed, state := rl.Allow()
		w.Header().Set("X-Rate-Limit-State", state.String())
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, "too many requests, state:", state)
			return
		}
		fmt.Fprintln(w, "ok, state:", state)
	})

	addr := ":8080"
	fmt.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("server error:", err)
	}
}
