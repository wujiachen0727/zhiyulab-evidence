# E5 实验记录

## 实验名
e5-options-vs-builder — Functional Options vs Builder 模式对比

## 对比结论
| 维度 | Builder 模式 | Functional Options |
|------|:-----------:|:-----------------:|
| 代码量 | ~40 行 | ~30 行（节省 25%）|
| 中间对象 | 需要 Builder 结构体 | 不需要 |
| 调用方式 | 链式 . 操作 | 函数参数传入 |
| 可扩展性 | 新增字段需加 WithXxx 方法 | 新增字段需加 WithXxx 函数 |
| 默认值 | 在 Builder 构造函数中设置 | 在 NewXxx 函数中设置 |
| Go 社区接受度 | 低（"Java in Go"） | 高（标准库广泛使用）|

## 关键洞察
Functional Options 的优势不仅是代码量少——它更符合 Go 的设计哲学：用函数类型替代接口，用可变参数替代链式调用。Builder 模式本身没有错，但在 Go 的语境下，Functional Options 是一种更"Go 式"的替代方案。

## 运行环境
- Go 版本：通过 go version 确认
- 运行时间：2026-06-08
