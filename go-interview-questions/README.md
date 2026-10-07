# 接得住追问的 10 道 Go 面试题 — 实验索引

配套文章：<https://www.wujiachen.com.cn/posts/go-interview-questions>

10 道经典 Go 面试题，每道题往下追一层，用可运行代码把那一层跑出来。这里放的是全部实验代码与原始输出。

## 环境

| 项 | 值 |
|----|-----|
| 工具链 | `go1.27.0 darwin/arm64` |
| 机器 | 14 逻辑核（`runtime.NumCPU() = 14`） |
| 第三方依赖 | 无，全部只使用 Go 标准库 |

## 复现方式

```bash
cd code
bash run-all.sh
```

脚本逐个进入每个实验目录执行 `go run .`，把 stdout 写入 `../output/{实验名}/result.txt`、stderr 写入 `../output/{实验名}/stderr.txt`。

**两点注意**：

1. `range-var-go121` / `go122` / `go126` 是三份**内容完全相同**的源码，唯一差别是 `go.mod` 里的 `go` 行。它们用来演示 `for range` 循环变量的语义由模块声明的语言版本决定，与工具链版本无关。
2. `gomaxprocs` 是计时类实验，绝对值会随机器负载浮动（实测波动约 ±3%），所以每档先热身一次、再取 3 次运行的最小值。仓库里的 `result.txt` 是文章引用的那一版；你自己重跑得到的绝对耗时会有小幅差异，但"加速比在协程数用尽处停止增长"这一结论稳定复现。

## 实验清单

| 目录 | 对应题目 | 论据 | 关键结论 |
|------|---------|:----:|---------|
| `code/make-vs-new` | 一、`make` 和 `new` | E1 | `new` 返回指针且指向 nil 引用类型；nil map 只有读安全（写 panic）；nil slice 可直接 `append` |
| `code/slice-append-alias` | 二、slice 和数组 | E2 | 子切片容量继承到原数组末尾——`b[:2]` 的 cap 是 3 不是 2，所以看似越界的追加仍共享内存；三下标 `c[:2:2]` 才能切断 |
| `code/typed-nil-interface` | 三、接口里的 nil | E6 | typed nil 返回后 `err == nil` 为 false；若 `Error()` 读字段，调用时直接 panic |
| `code/range-var-go121` / `-go122` / `-go126` | 四、`for range` 的变量 | E3 | 语义由 `go.mod` 的 `go` 行决定：`go 1.21` 共用变量，`go 1.22` 起每次迭代新变量 |
| `code/defer-semantics` | 五、`defer` | E4 | 参数在 defer 语句处求值；命名返回值可被改写、普通返回值不行；循环内 defer 到函数结束才执行 |
| `code/defer-loop-alloc` | 五、`defer`（第二半） | E4 | 循环外只有 `runtime.deferreturn`（开放编码）；循环内出现 `runtime.deferproc`（堆分配） |
| `code/map-concurrent` | 六、map 并发写 | E5 | 单协程写 nil map 是可 recover 的 panic；并发写是 `fatal error: concurrent map writes`，recover 接不住，退出码 2 |
| `code/goroutine-leak` | 七、goroutine 泄漏 | E7 | 两种泄漏形态分别 +100 / +50 协程，修复版本 +0；阻塞栈里同一调用帧重复出现 N 次 |
| `code/channel-close` | 八、关闭 channel | E8 | 关闭后接收返回零值且 `ok=false`；关闭后发送 panic；重复关闭 panic；nil channel 收发永久阻塞 |
| `code/context-cancel` | 九、`context` | E9 | 取消沿调用链传到第 5 层；不调用 cancel 则子 context 与其协程持续存活；WithTimeout 与取消同机制 |
| `code/gomaxprocs` | 十、`GOMAXPROCS` | E10 | 4 个 CPU 密集协程下 1/2/4/8 分别为 139/70/38/38 ms，加速比 1.00/1.98/3.65/3.67x |

## 论据统计

| 项 | 值 |
|----|-----|
| 独立论据 | 11 项（E1–E10 + 开篇的逻辑推演） |
| 表达手法 | 1 项（面试追问场景，不计入自造度） |
| 外部引用 | 3 处，均为官方发行说明或公开指南 |
| 自造论据占比 | 约 79% |

## 外部引用边界

文章只有三处外部引用，都不承载核心论点：

1. 招聘方视角的 Go 面试指南里"skip syntax trivia like `make` vs. `new`"一句——用作开篇旁证
2. Go 1.25 的容器感知 `GOMAXPROCS`——官方发行说明
3. Go 1.27 的泛型方法——官方发行说明

去掉这三处，文章的核心论点仍由 E1–E10 的全部实测输出支撑。

## 数据说明

- `data/experiments.md` 是正文引用数字的汇总表，每个数字都标注了对应的原始输出文件
- 全部数字为实测，无推演数据；正文引用的每个数字都能在 `output/` 里逐字找到
