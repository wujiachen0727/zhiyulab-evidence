# version-boundary 运行结果

**对应题目**：十、`GOMAXPROCS`（版本边界）
**目的**：确认"库 API 的可用性也由模块声明的语言版本决定"，用来支撑"背下来的 API 知识会随版本失效"这一论点
**运行方式**：分别进入 `code/version-boundary/go124` 与 `go125` 执行 `go run .`
**原始输出**：`go124.txt`、`go125.txt`

两个模块的 `main.go` **内容完全相同**，都调用 `sync.WaitGroup.Go`（Go 1.25 新增的方法），只在 `go.mod` 的 `go` 行上不同。

## 关键输出

`go 1.24` 模块（退出码 1）：

```
# vb124
./main.go:22:6: wg.Go undefined (type "sync".WaitGroup has no field or method Go)
```

`go 1.25` 模块（退出码 0）：

```
工具链 runtime.Version() = go1.25.11
runtime.NumCPU()         = 14
runtime.GOMAXPROCS(0)    = 14

wg.Go 汇总结果 = 6
```

## 结论

同一个方法调用，在 `go 1.24` 模块里是编译错误，在 `go 1.25` 模块里正常跑通。

这说明版本边界不只在语义层（比如 `for range` 的循环变量），还在**库 API 的可用性**上：模块声明的语言版本决定你能不能用某个标准库方法。把"怎么用 `WaitGroup`"这类答案背成固定版本，遇到旧模块会直接编译不过。

**注意**：`GOMAXPROCS` 默认值的容器感知差异需要 Linux 的 cgroup 环境才能观察到，本机（macOS）两版输出相同，故本组实验不覆盖那一项。容器部分是官方发行说明的结论，正文已注明。
