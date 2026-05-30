package main

import (
	"fmt"
	"reflect"
)

type UserID string
type OrderID string

func UniqueInts(input []int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(input))
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func UniqueStrings(input []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(input))
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func UniqueAny(input []any) []any {
	seen := map[any]bool{}
	out := make([]any, 0, len(input))
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func UniqueReflect(input any) (any, error) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %s", v.Kind())
	}
	seen := map[any]bool{}
	out := reflect.MakeSlice(v.Type(), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		if !seen[item] {
			seen[item] = true
			out = reflect.Append(out, v.Index(i))
		}
	}
	return out.Interface(), nil
}

func UniqueGeneric[T comparable](input []T) []T {
	seen := map[T]bool{}
	out := make([]T, 0, len(input))
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func sumIntsFromAny(values []any) (sum int, err any) {
	defer func() {
		if r := recover(); r != nil {
			err = r
		}
	}()
	for _, v := range values {
		// any 路径把类型检查推迟到这里；混入 string 后只能运行时炸。
		sum += v.(int)
	}
	return sum, nil
}

func main() {
	fmt.Println("[复制粘贴] int 去重:", UniqueInts([]int{1, 2, 2, 3}))
	fmt.Println("[复制粘贴] string 去重:", UniqueStrings([]string{"go", "java", "go"}))
	fmt.Println("[泛型] UserID 去重:", UniqueGeneric([]UserID{"u1", "u2", "u1"}))

	mixed := UniqueAny([]any{1, "2", 1})
	_, err := sumIntsFromAny(mixed)
	fmt.Println("[any] 混入 string 后的暴露阶段: runtime panic ->", err != nil)
	if err != nil {
		fmt.Println("[any] panic:", err)
	}

	// any 版本允许不同业务类型混入同一个函数
	userIDs := []any{UserID("u1"), OrderID("o1"), UserID("u2")}
	result := UniqueAny(userIDs)
	fmt.Println("[any] UserID/OrderID 混入:", result)

	reflected, reflectErr := UniqueReflect([]UserID{"u1", "u2", "u1"})
	fmt.Println("[reflect] UserID 去重:", reflected, "err:", reflectErr)
}
