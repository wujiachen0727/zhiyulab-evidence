package main

import (
	"errors"
	"fmt"
	"strings"
)

// 模拟批量操作中部分失败的场景
func batchCreateUsers(names []string) error {
	var errs []error
	for _, name := range names {
		if err := createUser(name); err != nil {
			errs = append(errs, fmt.Errorf("create user %q: %w", name, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

var (
	ErrDuplicateUser = errors.New("duplicate user")
	ErrInvalidName   = errors.New("invalid name")
)

func createUser(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	if name == "alice" {
		return ErrDuplicateUser
	}
	return nil
}

// ========================================
// 旧方式：手动拼接 fmt.Errorf
// ========================================
func batchCreateUsersOldStyle(names []string) error {
	var errMsgs []string
	for _, name := range names {
		if err := createUser(name); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("create user %q: %v", name, err))
		}
	}
	if len(errMsgs) > 0 {
		return errors.New(strings.Join(errMsgs, "; "))
	}
	return nil
}

func main() {
	names := []string{"alice", "", "bob", "charlie"}

	fmt.Println("=== errors.Join 方式（Go 1.20+）===")
	err := batchCreateUsers(names)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Is ErrDuplicateUser? %v\n", errors.Is(err, ErrDuplicateUser))
		fmt.Printf("Is ErrInvalidName? %v\n", errors.Is(err, ErrInvalidName))
	}

	fmt.Println("\n=== 手动拼接方式（旧写法）===")
	errOld := batchCreateUsersOldStyle(names)
	if errOld != nil {
		fmt.Printf("Error: %v\n", errOld)
		fmt.Printf("Is ErrDuplicateUser? %v\n", errors.Is(errOld, ErrDuplicateUser))
		fmt.Printf("Is ErrInvalidName? %v\n", errors.Is(errOld, ErrInvalidName))
	}

	fmt.Println("\n=== 关键差异 ===")
	fmt.Println("errors.Join：可以用 errors.Is 逐个检查错误类型 ✅")
	fmt.Println("手动拼接：errors.Is 无法穿透字符串拼接 ❌")
}
