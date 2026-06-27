# 实验报告：内联 vs 逃逸分析

> 2026-04-09 | Go 1.26.2 darwin/arm64

## 实验1：内联开关逃逸的对比

### 测试代码

```go
// 小函数：纯值类型
func add(a, b int) int {
    return a + b
}
```

### 逃逸分析结果

**内联开启**（默认）：
```
can inline add
inlining call to add
```

**内联关闭**（`-gcflags '-l'`）：
```
（无 can inline 输出）
```

### Benchmark 数据

| 函数 | 内联开启 (ns/op) | 内联关闭 (ns/op) | 差异 |
|------|:---:|:---:|:---:|
| `add` | 0.23 | 0.70 | **3x** |
| `doubleVal` | 0.78 | 0.72 | ~持平 |
| `addDirect`（基准） | 0.23 | 0.23 | 无差异 |

所有场景 0 B/op、0 allocs/op。

### 关键发现

1. **内联对纯值类型小函数的性能影响巨大**：`add` 内联后性能接近直接计算，不内联时慢 3 倍
2. **内联对逃逸分配的影响在简单场景下不显著**：Go 1.26 的逃逸分析足够智能，即使不内联也能判断值类型参数不逃逸
3. **性能差异主要来自函数调用开销**：栈帧创建、参数传递、返回值处理

## 实验2：二级逃逸分析日志

### `-gcflags '-m -m'` 输出摘要

**内联开启时**：
```
can inline distance with cost 22 as: func(...) float64 { ... }
can inline createAndSum with cost 30 as: func(int) int { ... }
cannot inline main: function too complex: cost 229 exceeds budget 80
```

**内联关闭时**：
```
（无 can inline 输出）
```

### 关键发现

1. **内联预算为 80**（Go 1.26）：cost 22 和 30 的函数被内联，cost 229 的 main 不被内联
2. **逃逸路径追踪**：`-m -m` 展示了完整的逃逸 flow chain，从 spill → slice-literal → call parameter
3. **fmt.Println 是逃逸大户**：所有传给 fmt.Println 的参数都逃逸到堆上

## 实验3：跨包调用

### 结果

Go 1.26 支持跨包内联，`internal.ProcessData` 被成功内联。跨包调用不再是逃逸的黑洞。

### 结论

现代 Go（1.26+）的编译器已经在很多方面比早期版本更智能。文章需要强调：**内联的真正价值不只是消除逃逸分配，更是消除函数调用开销**——3x 的性能差异来自调用开销，不是 GC 压力。
