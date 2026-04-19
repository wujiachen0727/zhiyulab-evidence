package main

import (
	"reflect"
	"testing"
)

type User struct {
	Name  string
	Age   int
	Email string
}

// 直接赋值
func BenchmarkDirectAssign(b *testing.B) {
	u := &User{}
	for i := 0; i < b.N; i++ {
		u.Name = "Alice"
		u.Age = 30
		u.Email = "alice@example.com"
	}
}

// 反射赋值
func BenchmarkReflectAssign(b *testing.B) {
	u := &User{}
	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(u).Elem()
		v.FieldByName("Name").SetString("Alice")
		v.FieldByName("Age").SetInt(30)
		v.FieldByName("Email").SetString("alice@example.com")
	}
}

// 反射赋值（缓存 FieldByIndex）
func BenchmarkReflectCachedAssign(b *testing.B) {
	u := &User{}
	t := reflect.TypeOf(*u)
	nameIdx, _ := t.FieldByName("Name")
	ageIdx, _ := t.FieldByName("Age")
	emailIdx, _ := t.FieldByName("Email")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(u).Elem()
		v.FieldByIndex(nameIdx.Index).SetString("Alice")
		v.FieldByIndex(ageIdx.Index).SetInt(30)
		v.FieldByIndex(emailIdx.Index).SetString("alice@example.com")
	}
}

// 直接方法调用
func (u *User) SetName(name string) { u.Name = name }

func BenchmarkDirectCall(b *testing.B) {
	u := &User{}
	for i := 0; i < b.N; i++ {
		u.SetName("Alice")
	}
}

// 反射方法调用
func BenchmarkReflectCall(b *testing.B) {
	u := &User{}
	args := []reflect.Value{reflect.ValueOf("Alice")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reflect.ValueOf(u).MethodByName("SetName").Call(args)
	}
}

// 接口+类型断言（反射的替代方案）
type Namer interface {
	SetName(string)
}

func BenchmarkInterfaceCall(b *testing.B) {
	var n Namer = &User{}
	for i := 0; i < b.N; i++ {
		n.SetName("Alice")
	}
}

// TypeOf 开销
func BenchmarkTypeOf(b *testing.B) {
	u := User{}
	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(u)
	}
}

// ValueOf 开销
func BenchmarkValueOf(b *testing.B) {
	u := User{}
	for i := 0; i < b.N; i++ {
		_ = reflect.ValueOf(u)
	}
}
