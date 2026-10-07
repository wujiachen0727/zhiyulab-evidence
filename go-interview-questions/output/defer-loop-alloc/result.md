# defer-loop-alloc 运行结果

**对应题目**：五、`defer`（第二半）
**目的**：确认循环内外的 `defer` 在编译结果上是否有差异
**运行方式**：`go build -gcflags=-S -o /dev/null .`（在 `code/defer-loop-alloc` 目录下）
**原始输出**：`asm.txt`

## 关键输出

```
main.nonLoop     CALL runtime.deferreturn(SB)

main.inLoop      CALL runtime.deferproc(SB)
                 CALL runtime.deferreturn(SB)
```

`nonLoop` 是循环外的单次 `defer`，`inLoop` 是循环内的 `defer`。

## 结论

循环外的 `defer` 走"开放编码"，defer 记录放在函数的堆栈里，只调用一次 `runtime.deferreturn`。

循环内的 `defer` 出现 `runtime.deferproc`：迭代次数编译期无法预知，编译器只能退回到堆分配。**所以循环里写 defer，除了把关闭操作推迟到函数结束，还在每一轮迭代上多付一次堆分配的代价。**
