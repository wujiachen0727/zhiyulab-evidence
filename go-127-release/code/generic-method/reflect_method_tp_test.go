package genericmethod

import (
	"reflect"
	"testing"
)

type Counter struct {
	n int
}

func (c Counter) Plain() int {
	return c.n
}

func (c Counter) Convert[U any](v U) U {
	return v
}

func TestReflectGenericMethodParams(t *testing.T) {
	typ := reflect.TypeOf(Counter{})
	n := typ.NumMethod()
	names := make([]string, 0, n)
	for i := 0; i < n; i++ {
		names = append(names, typ.Method(i).Name)
	}
	t.Logf("Counter NumMethod=%d names=%v", n, names)
}
