# 逃逸分析边界对照实验结果

[实测 Go 1.26.2 darwin/arm64 Apple M4 Pro]

## 编译器决策（`-gcflags="-m"`）

| 函数 | 变量/表达式 | 逃逸判定 |
|------|-----------|---------|
| returnByValue | x | ✅ does not escape |
| returnByPointer | x | ❌ moved to heap |
| closureCapture | func literal | ❌ escapes to heap |
| interfaceConvert | 42 | ❌ escapes to heap |
| sliceAlloc | make([]int, 0, 1) | ❌ escapes to heap |

## Benchmark 结果

| 场景 | ns/op | B/op | allocs/op | vs 栈分配 |
|------|------:|-----:|----------:|----------:|
| 栈分配（值返回） | ~0.74 | 0 | 0 | 1x |
| 堆分配（返回指针） | ~6.6 | 8 | 1 | ~9x |
| 堆分配（闭包捕获） | ~7.8 | 16 | 1 | ~10.5x |
| 接口转换 | ~0.75 | 0 | 0 | ~1x ⚠️ |
| 切片超容 | ~28.5 | 48 | 3 | ~38.5x |

## 关键发现

1. **返回指针 vs 值返回**：~9 倍差距，1 allocs/op vs 0 allocs/op——清晰的逃逸代价
2. **闭包捕获**：~10.5 倍差距，闭包需要额外 16B（函数指针 + 捕获变量）
3. **接口转换的意外**：编译器标记 `42 escapes to heap`，但 benchmark 显示 0 allocs/op。**这说明编译器的逃逸分析标记和运行时实际分配可能不一致**——Go 1.26 对小整数转 interface{} 有额外优化路径。这是一个重要细节：`-gcflags="-m"` 的输出是保守分析，不等于运行时行为。
4. **切片超容 append**：~38.5 倍差距，3 allocs/op（slice 本身 + 超容重分配 + append 到新 slice）
5. **内联对逃逸的影响**：未加 `//go:noinline` 时，所有函数都被内联，内联后逃逸分析消除了堆分配——这就是"优化链"的实例。

## 实验环境

- Go 版本：1.26.2
- 平台：darwin/arm64
- CPU：Apple M4 Pro
- 注意：所有函数加了 `//go:noinline` 防止内联干扰
