//go:build ignore

package main

type Getter interface {
	Value() int
}

type Box[T any] struct {
	v T
}

func (b Box[T]) Value() T {
	return b.v
}

func main() {
	var _ Getter = Box[int]{v: 1}
}
