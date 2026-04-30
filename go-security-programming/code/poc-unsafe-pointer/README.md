# E5: unsafe.Pointer 绕过 Go 类型封装

## 证伪结果

**假设**：unsafe.Pointer + reflect 能绕过 Go 的类型封装。  
**结果**：假设成立。即使所有字段小写（跨包不可见），
只要拿到 `*T` 指针，就能直接读写私有字段，包括"冻结"标志位等安全状态。

## 运行

```bash
go run main.go
```

## 观察要点

- Go 的类型安全是"编译期"和"常规 API"层面的，不是运行时内存隔离
- 威胁模型：依赖第三方库时，你无法保证对方不用 unsafe
- `reflect.Value.UnsafeAddr()` 加上 `unsafe.Pointer` 是绕过私有字段读写的标准手法

## 环境

Go 1.26.2 darwin/arm64

## 对应章节

第三章 第 4 个幻觉 · 小写字段=不可访问
