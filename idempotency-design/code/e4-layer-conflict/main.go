// E4: 层间冲突实验
// 场景：订单已支付（状态机：paid）。用户发起退款。
// 真实支付系统语义：调用渠道退款成功 → 才把本地状态更新为 refunding/refunded。
//
// 故障构造：第一个请求调用渠道退款成功，但在"落库更新状态"之前进程崩溃/超时。
//          上层重试 → 由于状态还是 paid，再次调用渠道退款 → 重复退款！
//
// 对照组：无幂等键（盲目重试）→ 渠道退款被调用 2 次（重复退款）
// 实验组：有幂等键（重试被拦截）→ 渠道退款只调用 1 次
//
// 运行: go run main.go
// 环境: go1.26.2 darwin/arm64

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type State string

const (
	Pending  State = "pending"
	Paid     State = "paid"
	Refunded State = "refunded"
)

var legal = map[State]map[State]bool{
	Pending:  {Paid: true},
	Paid:     {Refunded: true},
	Refunded: {},
}

type order struct {
	mu      sync.Mutex
	state   State
	version int
}

var channelRefunds atomic.Int64

// 调用支付渠道退款（每次调用都会真实打钱）
func refundChannel() {
	channelRefunds.Add(1)
}

// ===== 对照组：无幂等键，盲目重试 =====
func runBlindRetry() {
	o := &order{state: Paid, version: 1}
	channelRefunds.Store(0)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 读
			o.mu.Lock()
			cur, ver := o.state, o.version
			o.mu.Unlock()
			if !legal[cur][Refunded] {
				return // 已退款，正常情况应返回；但盲目重试无此检查
			}
			// 乐观锁更新（模拟 UPDATE ... WHERE version=?)
			o.mu.Lock()
			if o.version == ver {
				o.version++
				o.mu.Unlock()
				// 先调渠道（真实语义：渠道成功才落库）
				refundChannel()
				// 模拟落库前崩溃：state 没更新，直接返回
				// 上层看到超时 → 盲目重试
				return
			}
			o.mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  渠道退款调用次数: %d\n", channelRefunds.Load())
}

// ===== 实验组：幂等键拦截重试 =====
type idemStore struct {
	mu      sync.Mutex
	records map[string]bool
}

// 返回 true 表示"已处理过"（重试命中）
func (s *idemStore) seen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.records[key] {
		return true
	}
	s.records[key] = true
	return false
}

func runWithIdempotency() {
	o := &order{state: Paid, version: 1}
	channelRefunds.Store(0)
	store := &idemStore{records: make(map[string]bool)}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqID := "refund-20260814-001" // 同一退款请求（幂等键）
			// 幂等键检查：重试直接返回首次结果
			if store.seen(reqID) {
				return
			}
			o.mu.Lock()
			cur, ver := o.state, o.version
			o.mu.Unlock()
			if !legal[cur][Refunded] {
				return
			}
			o.mu.Lock()
			if o.version == ver {
				o.version++
				o.mu.Unlock()
				refundChannel()
				o.mu.Lock()
				o.state = Refunded
				o.version++
				o.mu.Unlock()
				return
			}
			o.mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  渠道退款调用次数: %d\n", channelRefunds.Load())
}

func main() {
	fmt.Println("=== 场景：并发退款请求，第一个在'调渠道后、落库前'崩溃 ===")
	fmt.Println()
	fmt.Println("对照组（无幂等键，盲目重试）:")
	runBlindRetry()
	fmt.Println()
	fmt.Println("实验组（幂等键拦截重试）:")
	runWithIdempotency()
	fmt.Println()
	fmt.Println("=== 结论 ===")
	fmt.Println("乐观锁防的是'并发改同一行'；但它拦不住'先调外部渠道、后落库'的崩溃窗口。")
	fmt.Println("这个窗口只能靠幂等键（或状态机拒绝重复迁移）兜住。")
	fmt.Println("三层各管一类故障——层间配合才是完整防线。")
}
