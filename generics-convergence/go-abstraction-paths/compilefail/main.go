package main

type UserID string
type OrderID string

func UniqueGeneric[T comparable](input []T) []T {
	return input
}

func main() {
	// 这段代码故意不能编译：[]UserID 里混入 OrderID，泛型路径在编译期拦截。
	_ = UniqueGeneric([]UserID{UserID("u1"), OrderID("o1")})
}
