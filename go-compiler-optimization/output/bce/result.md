# BCE（边界检查消除）实测结果

[实测 Go 1.26.2 darwin/arm64 Apple M4 Pro]

## 编译器决策（`-gcflags="-d=ssa/check_bce"`）

| 函数 | BCE 结果 | 说明 |
|------|---------|------|
| sumWithBCE | ✅ 无边界检查 | `for i := 0; i < len(data); i++` → 编译器证明 i < len(data) |
| sumWithoutBCE | ❌ Found IsInBounds | `data[indices[i]]` → 编译器无法证明 indices[i] < len(data) |
| sumWithRange | ✅ 无边界检查 | `for _, v := range data` → range 遍历天然安全 |

## Benchmark 结果

| 场景 | ns/op | B/op | allocs/op | vs BCE版 |
|------|------:|-----:|----------:|:--------:|
| for 循环 + BCE 消除 | ~255 | 0 | 0 | 1x |
| 间接索引 + 无 BCE | ~262 | 0 | 0 | ~1.03x |
| range 遍历 + BCE 消除 | ~258 | 0 | 0 | ~1.01x |

## 关键发现

1. **BCE 在 ARM 上的收益很小**：仅 ~3% 性能差异。原因是 Apple M4 的分支预测器非常强，边界检查分支几乎 100% 预测命中，预测正确的分支开销接近零。
2. **但编译器仍然在做 BCE**：消除边界检查不只是性能——还减少了代码体积（不需要插入边界检查指令）。
3. **x86 平台可能有更大差异**：ARM 的条件执行指令（cbz/cbnz）比 x86 的分支更高效，BCE 在 x86 上的收益可能更显著。
4. **注意**：`sumWithoutBCE` 还多了一次 `indices[i]` 的内存访问，~3% 的差异不完全来自边界检查，还包含了额外的间接访问开销。

## 实验环境

- Go 版本：1.26.2
- 平台：darwin/arm64
- CPU：Apple M4 Pro
- 注意：ARM 平台分支预测强，BCE 收益比 x86 小
