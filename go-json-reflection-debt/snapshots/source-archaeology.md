# 源码考古笔记

> 来源：Go 1.26.2 源码 `/opt/homebrew/Cellar/go/1.26.2/libexec/src/encoding/json/`
> 考古日期：2026-05-31
> 标注：[实测 Go 1.26.2]

## 1. v1 的缓存机制（encoding/json/encode.go）

### encoderCache — 编码器缓存

```go
// encode.go:379
var encoderCache sync.Map // map[reflect.Type]encoderFunc
```

**机制**：每次 `json.Marshal` 调用 `typeEncoder(t)`，先查 `sync.Map`。首次访问时用 `sync.OnceValue` 处理递归类型，防止死锁。

**暗债形态**：首次编码某类型时，必须完整走一遍 reflect 逻辑构建 encoderFunc。后续调用虽然命中缓存，但 `reflect.ValueOf` 调用仍然不可避免——每次 Marshal 都要付这个钱。

### cachedTypeFields — 字段缓存

```go
// encode.go:1329-1337
var fieldCache sync.Map // map[reflect.Type]structFields

func cachedTypeFields(t reflect.Type) structFields {
    if f, ok := fieldCache.Load(t); ok {
        return f.(structFields)
    }
    f, _ := fieldCache.LoadOrStore(t, typeFields(t))
    return f.(structFields)
}
```

**机制**：`typeFields` 执行 BFS 遍历所有嵌入字段，解析 struct tag，构建字段索引。代价高但只执行一次。

**暗债形态**：首次调用是"初始化锐刺"——30 字段 struct 的 `typeFields` 内部创建 map、slice、执行字符串操作，全部分配在堆上。

### typeFields 内部的 BFS 遍历

```go
// encode.go:1093-1140（关键段）
func typeFields(t reflect.Type) structFields {
    current := []field{}
    next := []field{{typ: t}}
    var count, nextCount map[reflect.Type]int
    visited := map[reflect.Type]bool{}
    var fields []field
    // BFS 循环...
    for i := 0; i < f.typ.NumField(); i++ {
        sf := f.typ.Field(i) // <- 每个字段一次 reflect 调用
        tag := sf.Tag.Get("json") // <- 字符串解析
    }
}
```

**暗债形态**：30 字段 = 至少 30 次 `reflect.Type.Field()` + 30 次 `StructTag.Get()`。每次 `Field()` 返回的 `StructField` 包含字符串（Name/Tag），可能逃逸到堆。

## 2. v2 的分层架构（encoding/json/v2/ + encoding/json/jsontext/）

### 核心改进：语法层与语义层分离

v2 拆成两个包：
- `encoding/json/jsontext`：纯语法处理（Token/Value 的读写），**不涉及 reflect**
- `encoding/json/v2`：语义映射（Go 值 ↔ JSON），使用 reflect 但优化了缓存

**关键洞察**：jsontext 层处理 90% 的字节操作（扫描、转义、缩进），这一层完全不需要 reflect。v1 把这两层混在一起，reflect 的"传染性"比实际需要的范围大得多。

### v2 的缓存优化

```go
// v2/arshal.go:540-553
var lookupArshalerCache sync.Map // map[reflect.Type]*arshaler

func lookupArshaler(t reflect.Type) *arshaler {
    if v, ok := lookupArshalerCache.Load(t); ok {
        return v.(*arshaler)
    }
    fncs := makeDefaultArshaler(t)
    fncs = makeMethodArshaler(fncs, t)
    fncs = makeTimeArshaler(fncs, t)
    v, _ := lookupArshalerCache.LoadOrStore(t, fncs)
    return v.(*arshaler)
}
```

**对比 v1**：v2 的 arshaler 是一个合并的结构，包含 marshal + unmarshal 函数。v1 分别缓存 encoder 和 decoder，两次查找。

### v2 的 sync.Pool 使用

```go
// v2/arshal.go:556
var stringsPools = &sync.Pool{New: func() any { return new(stringSlice) }}
```

v2 对临时切片使用 sync.Pool 复用，减少 Unmarshal 时的分配。这是 v1 Unmarshal 23 allocs → v2 3 allocs 的关键原因之一。

## 3. 数据总结

| 指标 | v1 | v2 | 减少比例 |
|------|----|----|:--------:|
| Unmarshal allocs/op | 23 | 3 | **87%** |
| Unmarshal ns/op | 4192 | 1769 | **58%** |
| Marshal allocs/op | 1 | 1 | 0% |
| Marshal ns/op | 788 | 1079 | +37%（v2 Marshal 略慢） |

**关键发现**：
1. v2 的主要优化在 Unmarshal（反序列化），因为 v1 的 Unmarshal 路径有大量反射驱动的分配
2. v2 Marshal 略慢于 v1，推测是 v2 的 jsontext 分层引入了额外间接调用
3. 87% 的 alloc 减少证明：不用 codegen，仅靠更智能的运行时策略（缓存合并 + Pool 复用 + 语法层分离），就能大幅降低反射的堆分配代价
