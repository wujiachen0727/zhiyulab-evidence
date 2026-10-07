# 实测数据汇总

> 全部数据来自 `evidence/output/` 的真实运行输出，非估算、非推演。
> 环境：`go1.27.0 darwin/arm64`，14 核，macOS。E3 的工具链差异见备注。

## E1 `make` / `new`（`output/make-vs-new/result.txt`）

| 表达式 | 类型 | 结果 |
|--------|------|------|
| `new(map[string]int)` | `*map[string]int` | 指向的 map 是 nil |
| `new([]int)` | `*[]int` | 指向的 slice 是 nil |
| 读 `nilMap["missing"]` | int | `0`（不 panic） |
| `var m map[string]int; m["a"] = 1` | — | `panic: assignment to entry in nil map` |
| `(*new(map[string]int))["a"] = 1` | — | `panic: assignment to entry in nil map` |
| `var s []int; s = append(s, 1, 2)` | []int | `len=2 cap=2 值=[1 2]` |
| `make([]int, 0, 4)` | []int | `len=0 cap=4 是 nil: false` |

## E2 slice 与 append（`output/slice-append-alias/result.txt`）

| 场景 | 操作 | 结果 |
|------|------|------|
| 组 1 | `a := make([]int,3,4)`；`head := a[:2]`；`head = append(head, 99)` | `head` 与 `a` 都是 `[1 2 99]` |
| 组 2 | `b := make([]int,3,3)`；`b[:2]` 的 cap | **cap=3**（不是 2），追加后 `b` 也被改成 `[1 2 99]` |
| 组 2 | `c := make([]int,3,3)`；`c[:2:2]` 的 cap | **cap=2**，追加后重新分配，`c` 保持 `[1 2 3]` |
| 组 3 | `e := append(d,10)`；`f := append(d,20)` | 两者都是 `[1 2 3 20]`；改 `e[0]=777` 后 `f` 变成 `[777 2 3 20]` |
| 组 4 | `h := append(g, 10, 20)`（越界） | `h` 的 cap=8；改 `h[0]` 不影响 `g` |

**精确条件**：`append` 后总长度不超过容量时才写进共享数组。子切片的容量继承到原数组末尾，`a[:2]` 的容量通常是 `len(a)`；要让子切片彻底独立，得用三下标 `a[:2:2]`。

## E3 `for range` 循环变量（`output/range-var/`）

同一份源码，三个模块只差 `go.mod` 里的 `go` 行：

| 模块 | `go` 行 | 实际工具链 | 闭包捕获 | `&i` 是否同一地址 | 协程读到的值 |
|------|--------|-----------|---------|-----------------|-------------|
| `range-var-go121` | `go 1.21` | go1.24.13 | `3 3 3` | 相同 | `[3 3 3]` |
| `range-var-go122` | `go 1.22` | go1.24.13 | `1 2 3` | 不同 | `[1 2 3]` |
| `range-var-go126` | `go 1.26` | go1.27.0 | `1 2 3` | 不同 | `[1 2 3]` |

**精确条件**：决定语义的是模块声明的语言版本，不是工具链版本。`go 1.21` 与 `go 1.22` 两档的输出完全相反（协程结果已排序以排除调度顺序干扰）。

## E4 `defer`（`output/defer-semantics/result.txt` + `output/defer-loop-alloc/asm.txt`）

| 场景 | 代码 | 结果 |
|------|------|------|
| 参数求值 | `x := 1; defer fmt.Println("x =", x); x = 100` | 输出 `x = 1` |
| 闭包求值 | `y := 1; defer func(){ fmt.Println("y =", y) }(); y = 200` | 输出 `y = 200` |
| 普通返回值 | `r := 50; defer func(){ r = 100 }(); return r` | 返回 `50` |
| 命名返回值 | `func namedReturn() (r int) { defer func(){ r = 100 }(); return 50 }` | 返回 `100` |
| 循环内 defer 顺序 | 3 次迭代各注册一个 defer | 循环先跑完，然后 `3 → 2 → 1` |

汇编对照（`go build -gcflags=-S`，go1.27.0）：

| 函数 | defer 位置 | 出现的运行时调用 |
|------|-----------|----------------|
| `main.nonLoop` | 循环外 | 只有 `runtime.deferreturn` |
| `main.inLoop` | 循环内 | `runtime.deferproc` + `runtime.deferreturn` |

**精确条件**：循环外的 defer 走开放编码（栈上），循环内因为迭代次数编译期不可知，只能退化到 `runtime.deferproc`（堆分配）。

## E5 map 并发写（`output/map-concurrent/`）

| 场景 | 结果 | 退出码 |
|------|------|:------:|
| 单协程写 nil map | `panic: assignment to entry in nil map`，被 recover 接住 | 0 |
| 8 协程各写 10 万次（每个协程内都装了 recover） | `fatal error: concurrent map writes`，**没有任何 recover 触发** | 2 |

**精确条件**：`recover` 能接住 `assignment to entry in nil map`，接不住 `concurrent map writes`——后者是运行时致命错误，运行时直接终止进程。

## E6 接口 nil（`output/typed-nil-interface/result.txt`）

| 表达式 | 结果 |
|--------|------|
| `runGood() == nil`（返回裸 nil） | `true` |
| `runSafe() == nil`（返回 typed nil 的 `*SafeErr`） | `false` |
| `runSafe()` 的动态类型 | `*main.SafeErr` |
| `runSafe()` 的动态值（经 `%v` 调用 `Error()`） | `boom(未初始化)` |
| `var e error; e == nil` | `true` |
| `e = runSafe(); e == nil` | `false` |
| `runFatal()` 的 `err.Error()`（方法读字段） | `panic: invalid memory address or nil pointer dereference` |

**精确条件**：接口值只有"类型和值都为 nil"时才等于 nil。返回 typed nil 指针会让调用方走进错误分支；若 `Error()` 方法解引用了接收者的字段，调用时还会直接崩。

## E7 goroutine 泄漏（`output/goroutine-leak/result.txt`）

| 场景 | 调用次数 | 协程数变化 |
|------|:--------:|:----------:|
| `leakBySend()`（无接收者） | 100 | **+100** |
| `leakByContext()`（派生后不 cancel） | 50 | **+50** |
| `fixedByBuffer()`（缓冲 1） | 100 | **+0** |
| `fixedByCancel()`（显式 cancel） | 100 | **+0** |

阻塞栈统计：`main.leakBySend.func1()` 出现 100 次，`main.leakByContext.func1()` 出现 50 次。

**精确条件**：泄漏不产生任何错误输出，只是协程数单调增长；线上表现为同一个栈帧在 goroutine profile 里重复出现成千上万次。

## E8 channel 关闭（`output/channel-close/result.txt`）

| 场景 | 结果 |
|------|------|
| 关闭后第 1、2 次接收（缓冲里还有值） | `值=1 ok=true`、`值=2 ok=true` |
| 关闭后第 3、4 次接收 | `值=0 ok=false`、`值=0 ok=false` |
| 关闭后发送 | `panic: send on closed channel` |
| 重复关闭 | `panic: close of closed channel` |
| nil channel 接收（配 150ms 超时） | 无任何进展 |
| nil channel 发送（配 150ms 超时） | 无任何进展 |
| 接收方先 close，发送方随后发送 | 发送方 `panic: send on closed channel` |

**精确条件**：关闭只影响"读"的语义（返回零值 + `ok=false`），不影响"写"——写会直接 panic。接收方关闭 channel 会把发送方炸掉。

## E9 `context`（`output/context-cancel/result.txt`）

| 场景 | 结果 |
|------|------|
| 取消沿 5 层调用链传播 | 最底层收到取消，`ctx.Err() = context canceled`；从取消到底层退出 50ms（取整到 10ms，该值来自 50ms 定时器加调度抖动，精确到毫秒不可复现） |
| 派生 100 个子 context 但不 cancel | 协程数 `1 → 101` |
| 未取消的子 context | `Done()` 未关闭，`Err() = <nil>` |
| 调用 `cancel()` 之后 | `Done()` 已关闭，`Err() = context canceled` |
| `WithTimeout(80ms)` 到期 | `Err() = context deadline exceeded` |
| 值沿调用链传递 | 第 3 层拿到 `req-7f3a` |

## E10 `GOMAXPROCS`（`output/gomaxprocs/result.txt`）

环境：`runtime.NumCPU() = 14`，`runtime.GOMAXPROCS(0) = 14`。
测试：4 个 CPU 密集协程，每个 1.5 亿次累加。**每一档先热身一次，再取 3 次运行的最小值**——单次计时受调度抖动与机器负载影响，取最小值是这一档设置下最接近真实下界的稳定估计。

| GOMAXPROCS | 墙钟时间 | 相对 1 的加速比 |
|:----------:|:--------:|:--------------:|
| 1 | 139 ms | 1.00x |
| 2 | 70 ms | 1.98x |
| 4 | 38 ms | 3.65x |
| 8 | 38 ms | 3.67x |

**精确条件**：`GOMAXPROCS` 控制同时执行 Go 代码的操作系统线程数。协程数超过 `GOMAXPROCS` 后加速比停止增长（4 → 8 只多出 0.02x，因为只有 4 个协程可跑）。

**复现说明**：绝对值会随机器负载浮动（实测波动范围约 ±3%），但"在协程数用尽处停止增长"这一形状稳定复现。
