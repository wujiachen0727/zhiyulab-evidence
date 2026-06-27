# 接口去虚化对照实验结果

[实测 Go 1.26.2 darwin/arm64 Apple M4 Pro]

## 编译器决策

### 场景1：具体类型赋值给接口变量
- `devirtualizing p.Process to fastProcessor` ✅
- `inlining call to fastProcessor.Process` ✅
- 编译器看到 `var p Processor = fastProcessor{}`，确定具体类型，去虚化+内联

### 场景2：通过 slice 间接获取接口值
- `p.Process(i)` 无去虚化 ❌
- 无内联 ❌
- 编译器看到 `processors[0]`，无法确定运行时具体类型，必须动态分发

### 场景3：直接调用具体类型方法（对照）
- `inlining call to fastProcessor.Process` ✅
- 自然内联，无接口开销

## Benchmark 结果

| 场景 | ns/op | B/op | allocs/op | 去虚化 | 内联 | vs 基准 |
|------|------:|-----:|----------:|:-----:|:----:|:-------:|
| 直接调用具体类型 | ~259 | 0 | 0 | N/A | ✅ | 1x |
| 接口（可去虚化） | ~259 | 0 | 0 | ✅ | ✅ | ~1x |
| 接口（不可去虚化） | ~1273 | 0 | 0 | ❌ | ❌ | ~4.9x |

## 关键发现

1. **去虚化让接口调用"免费"**：场景1和场景3性能完全一致（~259ns），说明 Go 1.26 的去虚化能力很强——编译器能确定具体类型时，接口调用零开销。
2. **不可去虚化 = ~5 倍性能损失**：当编译器无法确定具体类型时，动态分发的开销约 5 倍。
3. **allocs/op 都是 0**：性能差异完全来自分发方式（间接跳转 + 无法内联），不是 GC 开销。
4. **Go 1.26 的去虚化条件**：编译器需要看到接口变量的具体类型赋值。如果类型通过 slice/map/channel 间接获取，编译器放弃去虚化。

## 实验环境

- Go 版本：1.26.2
- 平台：darwin/arm64
- CPU：Apple M4 Pro
