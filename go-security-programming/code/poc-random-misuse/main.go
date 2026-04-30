// Package main demonstrates the classic math/rand vs crypto/rand misuse.
//
// PoC E3: "Go 编译器让 math/rand 和 crypto/rand 看起来一样——
// 但一个是可预测的伪随机，一个是密码学安全。业务代码里混用即漏洞。"
//
// Run: go run main.go
package main

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
)

// --- 危险版本：用 math/rand 生成 session token ---

// generateSessionTokenUnsafe 看起来完全正常的 token 生成函数。
// Go 1.22+ math/rand/v2 默认使用运行时随机种子（不再是固定种子），
// 看起来"更安全"——但它**仍然不是密码学安全的**。
// 攻击者如果能观察到若干个输出，可以恢复内部状态、预测后续输出。
func generateSessionTokenUnsafe() string {
	buf := make([]byte, 16)
	for i := range buf {
		buf[i] = byte(mrand.IntN(256))
	}
	return hex.EncodeToString(buf)
}

// --- 安全版本：用 crypto/rand ---

func generateSessionTokenSafe() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		// 在 Unix 系统上 crypto/rand 几乎不会失败
		panic(err)
	}
	return hex.EncodeToString(buf)
}

// --- 证伪：演示 math/rand 的内部状态可恢复 ---

// demonstratePredictability 不做"实际的状态恢复攻击"——那需要专门的 PCG 反演代码。
// 但我们可以展示：如果两个进程用了相同的种子，输出完全一致。
// 现实威胁模型是：攻击者拿到一个已泄漏的 token，如果种子是可猜的
// （比如服务重启时用了 time.Now().Unix() 做种子），就能预测。
func demonstratePredictability() {
	// 两个 PCG 源用完全相同的种子——输出会一模一样
	src1 := mrand.NewPCG(42, 1024)
	src2 := mrand.NewPCG(42, 1024)
	r1 := mrand.New(src1)
	r2 := mrand.New(src2)

	fmt.Println("两个相同种子的 math/rand 源的输出：")
	for i := 0; i < 5; i++ {
		v1 := r1.IntN(1_000_000)
		v2 := r2.IntN(1_000_000)
		fmt.Printf("  第 %d 次: r1=%d, r2=%d  (相同? %v)\n", i+1, v1, v2, v1 == v2)
	}
	fmt.Println()
	fmt.Println("含义：如果攻击者能猜到/泄漏种子（如 time.Now().Unix()），")
	fmt.Println("       就能完全复现你生成的所有 token。")
}

func main() {
	fmt.Println("=== 看起来都是 16 字节随机 token ===")
	fmt.Printf("  math/rand 版本:   %s\n", generateSessionTokenUnsafe())
	fmt.Printf("  crypto/rand 版本: %s\n", generateSessionTokenSafe())
	fmt.Println()
	fmt.Println("从肉眼看不出差异。代码 review 时也看不出——Go 编译器不会警告。")
	fmt.Println("gosec 会警告（规则 G404），但前提是你跑了 gosec。")
	fmt.Println()

	fmt.Println("=== 证伪：math/rand 的确定性 ===")
	demonstratePredictability()

	fmt.Println("=== 结论 ===")
	fmt.Println("1. 编译器无法区分 math/rand 和 crypto/rand 的调用场景")
	fmt.Println("2. Go 标准库命名上很诚实——math vs crypto——但业务代码经常就手误或手懒")
	fmt.Println("3. 任何用于安全目的的随机（token/salt/nonce/key）必须 crypto/rand")
	fmt.Println("4. math/rand/v2 虽然默认用系统随机种子，但算法仍非密码学安全")
}
