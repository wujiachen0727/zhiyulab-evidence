package genericmethod

import (
	"reflect"
	"testing"
)

type Box[T any] struct {
	v T
}

func (b Box[T]) Value() T {
	return b.v
}

func (b Box[T]) Plain() int {
	return 1
}

func TestReflectMethodSet(t *testing.T) {
	typ := reflect.TypeOf(Box[int]{})
	n := typ.NumMethod()
	names := make([]string, 0, n)
	for i := 0; i < n; i++ {
		names = append(names, typ.Method(i).Name)
	}
	t.Logf("NumMethod=%d names=%v", n, names)
}
