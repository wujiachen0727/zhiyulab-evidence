package main

import (
	"errors"
	"fmt"
)

// SafeErr 的 Error 方法不解引用任何字段，nil 接收者也能调用
type SafeErr struct{ code int }

func (e *SafeErr) Error() string {
	if e == nil {
		return "boom(未初始化)"
	}
	return fmt.Sprintf("boom(%d)", e.code)
}

// FatalErr 的 Error 方法会读字段，nil 接收者调用时崩溃
type FatalErr struct{ code int }

func (e *FatalErr) Error() string { return fmt.Sprintf("fatal(%d)", e.code) }

// runSafe 返回"类型为 *SafeErr、值为 nil"的指针
func runSafe() error {
	var e *SafeErr
	return e
}

// runFatal 返回"类型为 *FatalErr、值为 nil"的指针
func runFatal() error {
	var e *FatalErr
	return e
}

// runGood 成功路径返回裸 nil
func runGood() error { return nil }

func main() {
	fmt.Println("== 1. 返回值与 nil 的比较 ==")
	fmt.Printf("  runGood()  == nil            -> %v\n", runGood() == nil)
	fmt.Printf("  runSafe()  == nil            -> %v\n", runSafe() == nil)
	fmt.Printf("  runSafe()  的动态类型         -> %T\n", runSafe())
	fmt.Printf("  runSafe()  的动态值           -> %v\n", runSafe())

	fmt.Println()
	fmt.Println("== 2. 调用方会因此走错分支 ==")
	if err := runSafe(); err != nil {
		fmt.Printf("  调用方判定为出错，错误消息：%q\n", err.Error())
	} else {
		fmt.Println("  调用方判定为成功")
	}

	fmt.Println()
	fmt.Println("== 3. 接口值什么时候才是 nil ==")
	var e error
	fmt.Printf("  声明但未赋值          e == nil -> %v\n", e == nil)
	e = runSafe()
	fmt.Printf("  赋了 typed nil        e == nil -> %v\n", e == nil)
	e = &SafeErr{code: 7}
	fmt.Printf("  赋了真实错误          e == nil -> %v\n", e == nil)

	fmt.Println()
	fmt.Println("== 4. 更糟的一种：Error 方法一调用就崩 ==")
	if err := runFatal(); err != nil {
		fmt.Println("  调用方判定为出错，接着调用 err.Error() 打日志……")
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  err.Error() 触发 panic：%v\n", r)
				}
			}()
			_ = err.Error()
		}()
	}

	fmt.Println()
	fmt.Println("== 5. 识别 typed nil 的两种方式 ==")
	if err := runSafe(); err != nil {
		var target *SafeErr
		if errors.As(err, &target) && target == nil {
			fmt.Println("  errors.As 取出后 target == nil，可以判定这是个空指针错误")
		} else {
			fmt.Printf("  errors.As 判定：target 非空，err = %v\n", err)
		}
	}
}
