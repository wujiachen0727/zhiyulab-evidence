# 内联优化边界对照实验结果

[实测 Go 1.26.2 darwin/arm64 Apple M4 Pro]

## 编译器决策

### 内联友好版（直接调用小函数）
- `add` 函数：✅ can inline
- `callAdd` 函数：✅ can inline
- `callAdd` 内的 `add(i, i+1)` 调用：✅ inlining call to add

### 接口调用版（动态分发，无法内联）
- `dynamicInterfaceCall` 函数：❌ 不在 can inline 列表中（函数内有 slice 索引+接口调用，超过内联预算）
- `calc.Calculate(i, i+1)` 调用：❌ 未去虚化，未内联

## Benchmark 结果

| 场景 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| 直接调用（内联友好） | ~262 | 0 | 0 |
| 接口调用（不可内联） | ~1478 | 0 | 0 |
| **性能差距** | **~5.6x** | — | — |

## 关键发现

1. **性能差异来自分发方式，不是 GC**：两者 allocs/op 都是 0，证明差异来自动态分发开销（接口调用的间接跳转 + 无法内联），而非堆分配。
2. **同样的加法逻辑，编译器"看得懂"vs"看不懂"导致 5.6 倍差距**。
3. **Go 1.26 的去虚化能力**：当编译器能确定接口的具体类型时（如 `calc := concreteAdder{}` 后直接调用 `calc.Calculate()`），会自动去虚化并内联。但当类型在运行时才确定（如从 slice 中取出），编译器无法去虚化。

## 实验环境

- Go 版本：1.26.2
- 平台：darwin/arm64
- CPU：Apple M4 Pro
- 运行命令：`go test -bench=. -benchmem -count=5`
