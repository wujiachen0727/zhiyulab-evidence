// Package main demonstrates unsafe.Pointer + reflect breaking invariants.
//
// PoC E5: "类型系统保证你看不到的字段，unsafe.Pointer 能看到还能改。"
//
// Run: go run main.go
package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Account 模拟一个有"不变量"的账户对象：
// balance 不能直接被外部改——要走 Deposit/Withdraw。
// 这是 OO 里最基本的封装。
type Account struct {
	owner   string
	balance int64 // 私有字段，按契约只能通过方法修改
	frozen  bool  // 风控标志，冻结时拒绝出金
}

func NewAccount(owner string, balance int64) *Account {
	return &Account{owner: owner, balance: balance}
}

func (a *Account) Deposit(amount int64) {
	if a.frozen {
		return
	}
	a.balance += amount
}

func (a *Account) Withdraw(amount int64) bool {
	if a.frozen || a.balance < amount {
		return false
	}
	a.balance -= amount
	return true
}

func (a *Account) Freeze() { a.frozen = true }
func (a *Account) Balance() int64 { return a.balance }
func (a *Account) Owner() string  { return a.owner }

// --- 攻击代码 ---

// bypassBalance 演示 unsafe.Pointer + reflect 绕过封装。
// 只要拿到 *Account 指针（比如某个 RPC 传过来的参数，或某个配置加载器的返回值），
// 就可以不调用任何方法、直接改内存。
func bypassBalance(a *Account, newBalance int64) {
	// 通过 reflect 拿字段偏移
	t := reflect.TypeOf(*a)
	field, _ := t.FieldByName("balance")

	// 通过 unsafe.Pointer 拿到字段起始地址
	balancePtr := (*int64)(unsafe.Pointer(uintptr(unsafe.Pointer(a)) + field.Offset))
	*balancePtr = newBalance
}

// unfreezeViaReflect 用反射 + unsafe 强行解冻。
func unfreezeViaReflect(a *Account) {
	v := reflect.ValueOf(a).Elem()
	frozen := v.FieldByName("frozen")

	// reflect 默认不让改私有字段——但加 unsafe.Pointer 绕一下就行
	ptr := unsafe.Pointer(frozen.UnsafeAddr())
	*(*bool)(ptr) = false
}

func main() {
	a := NewAccount("alice", 1000)
	a.Freeze()

	fmt.Println("=== 正常路径：封装起效 ===")
	fmt.Printf("  初始余额: %d, 冻结: %v\n", a.Balance(), a.frozen)
	ok := a.Withdraw(500)
	fmt.Printf("  Withdraw(500) 结果: %v（冻结状态拒绝出金）\n", ok)
	fmt.Printf("  余额: %d\n\n", a.Balance())

	fmt.Println("=== unsafe.Pointer 绕过封装 ===")
	bypassBalance(a, 999_999_999)
	fmt.Printf("  bypassBalance 后余额: %d（绕过 Deposit/Withdraw）\n", a.Balance())

	unfreezeViaReflect(a)
	fmt.Printf("  unfreezeViaReflect 后冻结: %v\n", a.frozen)
	ok = a.Withdraw(100_000_000)
	fmt.Printf("  Withdraw(1亿) 结果: %v（冻结被绕过，余额也是伪造的）\n", ok)
	fmt.Printf("  最终余额: %d\n\n", a.Balance())

	fmt.Println("=== 关键观察 ===")
	fmt.Println("1. Account 结构体所有字段小写——Go 类型系统禁止跨包访问")
	fmt.Println("2. 但只要拿到 *Account 指针，unsafe.Pointer 能读写任何字段")
	fmt.Println("3. reflect 默认保护私有字段，unsafe.Pointer(v.UnsafeAddr()) 绕过")
	fmt.Println("4. Go 的类型安全是'编译期'和'常规 API'的——不是运行时内存隔离")
	fmt.Println("5. 威胁模型：依赖第三方库时，你无法保证对方不用 unsafe")
}
