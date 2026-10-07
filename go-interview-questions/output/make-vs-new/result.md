# make-vs-new 运行结果

**对应题目**：一、`make` 和 `new`
**目的**：确认 `new` 的"置零值"对引用类型意味着什么，以及 nil map 与 nil slice 的可操作性差异
**运行方式**：`cd code/make-vs-new && go run .`
**原始输出**：`result.txt`（stdout）

## 关键输出

```
new(map[string]int)    类型 *map[string]int，指向的 map 是 nil: true
读 nil map 的键:        0
向 nil map 写入                       panic: assignment to entry in nil map
向 *new(map[string]int) 写入          panic: assignment to entry in nil map
var s []int          len=0 cap=0 是 nil: true
append 之后           len=2 cap=2 值=[1 2]
```

## 结论

`new` 把内存置成零值，对 map 来说就是返回一个指向 nil map 的指针。nil map 只有读是安全的——读不存在的键返回零值，写会 panic。slice 相反，nil slice 可以直接 `append`，`append` 自己会分配底层数组。想拿到可用的 map，得用 `make`。
