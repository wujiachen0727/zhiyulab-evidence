package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// eface 是 Go 运行时中 interface{} 的底层结构
// 任何赋值给 interface{} 的值都会被打包成 (type, data) 二元组
type eface struct {
	_type unsafe.Pointer // 指向类型信息
	data  unsafe.Pointer // 指向实际数据
}

func main() {
	// 一个普通的 int 值
	x := 42

	// 装箱：赋值给 interface{} 时，Go 自动打包为 (type, data) 二元组
	var i interface{} = x

	// 用 unsafe 窥探 interface{} 的内存布局
	ef := (*eface)(unsafe.Pointer(&i))
	fmt.Println("=== interface{} 的内存真相 ===")
	fmt.Printf("interface{} 本身大小: %d 字节（两个指针）\n", unsafe.Sizeof(i))
	fmt.Printf("type 指针: %p\n", ef._type)
	fmt.Printf("data 指针: %p\n", ef.data)

	// reflect 所做的，就是把这个二元组"拆箱"
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	fmt.Println("\n=== reflect 拆箱结果 ===")
	fmt.Printf("reflect.TypeOf  → %v (Kind: %v)\n", t, t.Kind())
	fmt.Printf("reflect.ValueOf → %v (Type: %v)\n", v, v.Type())
	fmt.Printf("取回原始值     → %v\n", v.Interface())

	// 关键洞察：reflect.TypeOf 和 reflect.ValueOf 的输入就是 interface{}
	// 它们的函数签名是 func TypeOf(i any) Type 和 func ValueOf(i any) Value
	// 也就是说——你传给反射的东西，已经是一个二元组了
	// 反射只是帮你把这个二元组拆开来看

	fmt.Println("\n=== 对比：struct 的装箱和拆箱 ===")
	type User struct {
		Name string
		Age  int
	}
	u := User{Name: "Alice", Age: 30}
	var iu interface{} = u
	efu := (*eface)(unsafe.Pointer(&iu))
	fmt.Printf("User 装箱后: type=%p, data=%p\n", efu._type, efu.data)

	tu := reflect.TypeOf(iu)
	vu := reflect.ValueOf(iu)
	fmt.Printf("reflect.TypeOf  → %v (NumField: %d)\n", tu, tu.NumField())
	for j := 0; j < tu.NumField(); j++ {
		f := tu.Field(j)
		fmt.Printf("  Field[%d]: %s %v = %v\n", j, f.Name, f.Type, vu.Field(j))
	}
}
