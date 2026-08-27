//go:build ignore

package main

// Container 带类型参数的方法在 Go 1.27 可编译
type Container[T any] struct {
	v T
}

func (c Container[T]) Value() T {
	return c.v
}

func main() {
	_ = Container[int]{v: 1}.Value()
}
