# E3 Interface 拆箱实验结果

**环境**：Go 1.26.2, darwin/arm64
**实测时间**：2026-04-12
**标注**：[实测 Go 1.26.2]

## 实验输出

```
=== interface{} 的内存真相 ===
interface{} 本身大小: 16 字节（两个指针）
type 指针: 0x102463040
data 指针: 0x1023af9e8

=== reflect 拆箱结果 ===
reflect.TypeOf  → int (Kind: int)
reflect.ValueOf → 42 (Type: int)
取回原始值     → 42

=== 对比：struct 的装箱和拆箱 ===
User 装箱后: type=0x10246d140, data=0x10247f990
reflect.TypeOf  → main.User (NumField: 2)
  Field[0]: Name string = Alice
  Field[1]: Age int = 30
```

## 关键发现

1. **interface{} = 16 字节 = 两个指针**：一个指向类型信息（type），一个指向实际数据（data）
2. **reflect 的输入就是 interface{}**：`reflect.TypeOf(i any)` 和 `reflect.ValueOf(i any)` 的参数是 `any`（即 `interface{}`），所以你传给反射的东西在传入之前已经被"装箱"了
3. **反射不是魔法，是拆箱**：`reflect.TypeOf` 读取的就是 interface 二元组中的 type 指针，`reflect.ValueOf` 读取的就是 data 指针并包装为 `reflect.Value`
4. **struct 的字段信息也在 type 指针中**：通过 type 指针可以拿到 `NumField()`、`Field(i)` 等所有结构体元数据——这些信息在编译时就写入了可执行文件，反射只是在运行时读取它们
