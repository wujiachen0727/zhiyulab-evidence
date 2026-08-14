// E2: 状态机防乱序实验
// 模拟：支付回调乱序到达（网络重试/异步回调导致顺序颠倒）。
//
// 对照组 A：无状态机——回调直接改状态，乱序到达导致状态被"回拨"
//           （业务上应先支付成功、后发货；乱序时发货回调先到 → 直接改 → 已发货；
//            随后支付成功回调才到 → 直接改 → 已支付。订单显示"已支付"但实际已发货？）
// 实验组 B：有状态机——只允许合法迁移（pending→paid→shipped），非法迁移被拒绝
//
// 运行: go run main.go
// 环境: go1.26.2 darwin/arm64

package main

import "fmt"

type State string

const (
	Pending  State = "pending"  // 待支付
	Paid     State = "paid"     // 已支付
	Shipped  State = "shipped"  // 已发货
	Canceled State = "canceled" // 已取消
)

// 合法迁移表：状态机（单向不可回退）
var legalTransitions = map[State]map[State]bool{
	Pending:  {Paid: true, Canceled: true},
	Paid:     {Shipped: true},
	Shipped:  {},
	Canceled: {},
}

func valid(from, to State) bool {
	return legalTransitions[from][to]
}

// 业务序列：支付成功回调(→paid) 应在 发货回调(→shipped) 之前
// 乱序到达：发货回调先到（第1个），支付成功回调后到（第2个）
func main() {
	// ===== 场景 1：乱序回调 =====
	callbacks := []struct{ name string; to State }{
		{"发货回调", Shipped}, // 业务顺序上应该后到，但网络先到了
		{"支付成功回调", Paid},
	}

	fmt.Println("=== 场景1：支付回调乱序到达（业务顺序：先paid后shipped，实际先收到shipped）===")
	fmt.Println()

	// 对照组 A：无状态机，回调直接改状态
	stateA := Pending
	fmt.Println("对照组A（无状态机，直接改状态）:")
	for _, cb := range callbacks {
		fmt.Printf("  收到%s→%s: %s → %s\n", cb.name, cb.to, stateA, cb.to)
		stateA = cb.to
	}
	fmt.Printf("  >>> 最终状态: %s —— 订单先被改成了已发货（支付还没成功），\n", stateA)
	fmt.Printf("  >>>          随后支付回调又把状态改回已支付——状态来回跳，\n")
	fmt.Printf("  >>>          中间出现过'未支付却已发货'的错误状态。\n")
	fmt.Printf("  >>> 更糟的是：如果之后再来一个'取消回调'，状态能直接回拨到已取消。\n")
	fmt.Println()

	// 实验组 B：有状态机，非法迁移被拒
	stateB := Pending
	fmt.Println("实验组B（有状态机，只允许合法迁移）:")
	for _, cb := range callbacks {
		if valid(stateB, cb.to) {
			fmt.Printf("  收到%s→%s: %s → %s ✅ 合法迁移\n", cb.name, cb.to, stateB, cb.to)
			stateB = cb.to
		} else {
			fmt.Printf("  收到%s→%s: %s → %s ❌ 非法迁移被拒（订单保持 %s）\n", cb.name, cb.to, stateB, cb.to, stateB)
		}
	}
	fmt.Printf("  >>> 最终状态: %s —— 发货回调被拒（尚未支付不能发货）。\n", stateB)
	fmt.Printf("  >>> 乱序没有造成错误状态：等支付成功回调到达后，状态回到正确轨道。\n")
	fmt.Println()

	// ===== 场景 2：支付成功后，取消回调乱序到达 =====
	fmt.Println("=== 场景2：已支付后，取消回调到达（业务上不允许支付后退款取消？本例禁止）===")
	stateC := Paid
	if valid(stateC, Canceled) {
		fmt.Printf("  已支付→取消: ✅ 允许（实际业务可退款，本例简化禁止）\n")
	} else {
		fmt.Printf("  已支付→取消: ❌ 非法迁移被拒（订单保持 %s）——需要走退款流程，不是直接改状态\n", stateC)
	}

	// ===== 场景 3：重复回调（同状态重复到达）=====
	fmt.Println()
	fmt.Println("=== 场景3：支付成功回调重复到达 ===")
	stateD := Pending
	for i := 0; i < 3; i++ {
		if valid(stateD, Paid) {
			fmt.Printf("  第%d次paid回调: %s → %s ✅（第一次生效）\n", i+1, stateD, Paid)
			stateD = Paid
		} else {
			fmt.Printf("  第%d次paid回调: %s → %s ❌ 已处于paid，重复回调被忽略（幂等）\n", i+1, stateD, Paid)
		}
	}
}
