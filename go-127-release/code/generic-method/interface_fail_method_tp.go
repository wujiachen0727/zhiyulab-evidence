//go:build ignore

package main

// Counter 非泛型类型 + 方法自带类型参数 — 官方规则：不能实现接口
type Counter struct {
	n int
}

func (c Counter) Convert[U any](v U) U {
	return v
}

type IntConverter interface {
	Convert(int) int
}

func main() {
	var _ IntConverter = Counter{n: 1}
}
