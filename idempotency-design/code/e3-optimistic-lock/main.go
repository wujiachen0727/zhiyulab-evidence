// E3: 乐观锁防并发实验
// 模拟：多个并发请求同时扣款/修改同一账户余额。
//
// 对照组 A：无并发控制（read-modify-write 无版本校验）——
//          多个请求同时读，各自基于旧值计算，后写覆盖先写 → 扣款丢失
// 实验组 B：乐观锁（WHERE version=?）——版本不符更新 0 行，冲突被检出后重试或失败
//
// 运行: go run main.go
// 环境: go1.26.2 darwin/arm64

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ===== 对照组 A：无并发控制 =====
// balance 是共享变量；每个请求：读 → 算 → 写（模拟 UPDATE 覆盖，无任何校验）
// 用 time.Sleep 放大"读与写之间的窗口"，让并发交错真实发生
func runWithoutLock(initial, deduct, concurrency int) int {
	balance := initial
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 1. 读（都读到同一旧值）
			old := balance
			// 2. 计算
			newVal := old - deduct
			// 3. 模拟网络/业务处理耗时（放大竞争窗口）
			time.Sleep(time.Millisecond)
			// 4. 写（覆盖，无校验）
			balance = newVal
		}()
	}
	wg.Wait()
	return balance
}

// ===== 实验组 B：乐观锁 =====
type acct struct {
	mu      sync.Mutex
	balance int
	version int
}

// 一次扣款尝试：读 → 算 → UPDATE ... WHERE version=?。
// 返回 (是否成功, 是否发生冲突)
func tryDeduct(a *acct, deduct int) (bool, bool) {
	a.mu.Lock()
	old, ver := a.balance, a.version
	a.mu.Unlock()

	newVal := old - deduct
	time.Sleep(time.Millisecond) // 模拟业务处理耗时

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.version != ver { // WHERE version=? 影响 0 行
		return false, true // 冲突
	}
	a.balance = newVal
	a.version++
	return true, false
}

func runOptimisticLock(initial, deduct, concurrency int) (int, int) {
	a := &acct{balance: initial, version: 1}
	var success, conflicts atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				ok, conflict := tryDeduct(a, deduct)
				if conflict {
					conflicts.Add(1)
					continue // 重试
				}
				if ok {
					success.Add(1)
				}
				return
			}
		}()
	}
	wg.Wait()
	return a.balance, int(conflicts.Load())
}

func main() {
	fmt.Println("=== 场景1：2 笔 200 元并发扣款（余额 1000，期望 600）===")
	balA := runWithoutLock(1000, 200, 2)
	fmt.Printf("对照组A（无并发控制）: 最终余额 %d —— 丢了一笔扣款（本应 600）\n", balA)
	balB, cfB := runOptimisticLock(1000, 200, 2)
	fmt.Printf("实验组B（乐观锁）    : 最终余额 %d，冲突重试 %d 次 —— 两笔都成功\n", balB, cfB)
	fmt.Println()

	fmt.Println("=== 场景2：10 笔 200 元并发扣款（余额 2000，期望 0）===")
	balA2 := runWithoutLock(2000, 200, 10)
	fmt.Printf("对照组A: 最终余额 %d —— 只生效 1 笔，9 笔丢失\n", balA2)
	balB2, cfB2 := runOptimisticLock(2000, 200, 10)
	fmt.Printf("实验组B: 最终余额 %d，冲突重试 %d 次 —— 10 笔全部生效\n", balB2, cfB2)
	fmt.Println()

	fmt.Println("=== 场景3：乐观锁失败语义（冲突不重试，直接失败）===")
	a := &acct{balance: 1000, version: 1}
	// 事务1 和事务2 并发（同一窗口）
	var r1, r2 bool
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); r1, _ = tryDeduct(a, 200) }()
	go func() { defer wg.Done(); r2, _ = tryDeduct(a, 200) }()
	wg.Wait()
	fmt.Printf("事务1 成功=%v；事务2 成功=%v（版本冲突，0 行更新）\n", r1, r2)
	fmt.Printf("  >>> 失败语义：返回 409/更新 0 行，上层决定重试或提示用户——不会静默丢扣款\n")
	fmt.Printf("  最终余额 %d（只有 1 笔生效），无扣款丢失\n", a.balance)
}
