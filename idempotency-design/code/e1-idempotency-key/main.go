// E1: 幂等键防重复实验
// 模拟：客户端携带 Idempotency-Key 提交创建支付单请求。
// 对照组 A：无幂等键——同一请求重发 N 次，产生 N 个订单（重复入账）
// 实验组 B：有幂等键表（key 唯一约束）——同一 key 重发 N 次，只产生 1 个订单，后续返回首次结果
//
// 运行: go run main.go
// 环境: go1.26.2 darwin/arm64

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type order struct {
	ID     string
	Amount int
}

type idemRecord struct {
	orderID string
	amount  int
}

// 幂等键表
type idemStore struct {
	mu      sync.Mutex
	records map[string]idemRecord
}

func newIdemStore() *idemStore {
	return &idemStore{records: make(map[string]idemRecord)}
}

var orderSeq atomic.Int64

// 有幂等键的提交：key 已存在且参数一致 → 返回首次结果，不创建
// 参数不一致 → 返回错误（生产应拒绝，防止误用）
func (s *idemStore) submitWithKey(key string, amount int, orders *[]order) (string, bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec, ok := s.records[key]; ok {
		if rec.amount != amount {
			return rec.orderID, false, "PARAM_MISMATCH: 同 key 不同参数，服务端拒绝"
		}
		return rec.orderID, false, "REPLAY: 返回首次结果"
	}
	id := fmt.Sprintf("order-%d", orderSeq.Add(1))
	*orders = append(*orders, order{ID: id, Amount: amount})
	s.records[key] = idemRecord{orderID: id, amount: amount}
	return id, true, "CREATED"
}

// 无幂等键的提交：每次都是新订单
func submitWithoutKey(amount int, orders *[]order) string {
	id := fmt.Sprintf("order-%d", orderSeq.Add(1))
	*orders = append(*orders, order{ID: id, Amount: amount})
	return id
}

func main() {
	// ===== 场景：同一请求并发重发 100 次 =====
	const retries = 100

	// 对照组 A：无幂等键
	var ordersA []order
	var muA sync.Mutex
	var wgA sync.WaitGroup
	for i := 0; i < retries; i++ {
		wgA.Add(1)
		go func() {
			defer wgA.Done()
			muA.Lock()
			submitWithoutKey(998, &ordersA)
			muA.Unlock()
		}()
	}
	wgA.Wait()
	fmt.Printf("对照组A（无幂等键）: 重发 %d 次 → 订单数 %d（重复入账 %d 笔）\n",
		retries, len(ordersA), len(ordersA)-1)

	// 实验组 B：有幂等键（同一 key 并发重发）
	var ordersB []order
	orderSeq.Store(0) // B 组独立序号，便于阅读
	store := newIdemStore()
	var wgB sync.WaitGroup
	for i := 0; i < retries; i++ {
		wgB.Add(1)
		go func() {
			defer wgB.Done()
			store.submitWithKey("idem-20260814-001", 998, &ordersB)
		}()
	}
	wgB.Wait()
	fmt.Printf("实验组B（有幂等键）: 重发 %d 次 → 订单数 %d（重复入账 %d 笔）\n",
		retries, len(ordersB), len(ordersB)-1)

	// ===== 场景：幂等键返回首次结果（重试）=====
	id, created, status := store.submitWithKey("idem-20260814-001", 998, &ordersB)
	fmt.Printf("第 101 次重试同 key 同参数: 返回订单 %s, 新建=%v, 状态=%s\n", id, created, status)

	// ===== 场景：同 key 不同参数（误用防护）=====
	_, _, status2 := store.submitWithKey("idem-20260814-001", 999, &ordersB)
	fmt.Printf("同 key 不同金额: %s\n", status2)
}
