// 场景1：类型安全容器 —— 反射版 vs 泛型版
package scenario1

import "fmt"

// === 反射版（interface{} + type assertion）===

type ReflectStack struct {
	items []interface{}
}

func (s *ReflectStack) Push(item interface{}) {
	s.items = append(s.items, item)
}

func (s *ReflectStack) Pop() (interface{}, error) {
	if len(s.items) == 0 {
		return nil, fmt.Errorf("stack is empty")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

func (s *ReflectStack) Len() int {
	return len(s.items)
}

// === 泛型版 ===

type GenericStack[T any] struct {
	items []T
}

func (s *GenericStack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *GenericStack[T]) Pop() (T, error) {
	var zero T
	if len(s.items) == 0 {
		return zero, fmt.Errorf("stack is empty")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

func (s *GenericStack[T]) Len() int {
	return len(s.items)
}
