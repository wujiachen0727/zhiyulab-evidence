# Evidence 总索引 — go-sync-benchmark-showdown

> 全部论据自造（C 档，无真实经历，8/8 自造，自造度 100%）。
> 测量条件统一：Go 1.26.4 darwin/arm64（Apple M4 Pro，14 核），GOMAXPROCS=14，`go test -bench -benchmem`。
> 防 DCE：所有 benchmark 返回值赋给 package-level sink。
> 数据性质标注：下文 `[实测]` 为自跑 benchmark；`[源码推演]` 为标准库源码逻辑分析，非性能测量。

## 一、自造论据清单

| ID | 类型 | 描述 | 数据摘要 | 产出路径 |
|----|------|------|---------|---------|
| E1 | 实验验证 | Once vs Channel 单例获取开销 | Once 0.42ns；Channel 往返 16.4ns（≈39x）/ 带生产者 119.5ns（≈284x） | code/singleton, output/singleton |
| E2 | 实验验证 | atomic.AddInt64 vs Mutex 计数 | 无争用 1.1x；14 协程争用 atomic 19.5ns vs mutex 90.4ns（≈4.6x） | code/counter, output/counter |
| E3 | 实验验证 | WaitGroup 100/1000/10000 worker 退化 | 21.1μs / 234μs / 2.68ms（1000 比 100 慢 ≈11x） | code/waitgroup, output/waitgroup |
| E4 | 实验验证 | RWMutex vs Mutex 并发读 | 纯读 RWMutex 110.8ns 慢于 Mutex 70.1ns（1.6x）；读带工作量 RWMutex 109.3ns 略快于 130ns | code/rwmutex, output/rwmutex |
| E5 | 实验验证 | atomic.Value 无锁读 vs RWMutex 读 | 0.60ns vs 64.2ns（≈107x） | code/atomicvalue, output/atomicvalue |
| E6 | 实验验证 | errgroup vs WaitGroup 批量编排 | 2456ns vs 2473ns（≈持平，errgroup 多 2 allocs） | code/errgroup, output/errgroup |
| E7 | 源码推演 | sync.Once.doSlow 时序 + panic 陷阱 | 首次后零开销（done.Load 原子读）；f panic 后 done 仍置 1 → 永久失效 | 正文贴 once.go:67-80 |
| E8 | 源码推演 | Channel 单例慢根因 | chanrecv 走 c.lock + typedmemmove 拷贝；Once 仅原子读，无锁无拷贝 | 正文贴 chan.go:524 + once.go |

## 二、与源文章口述数字的偏差（诚实记录）

| 源文章声称 | 本文实测 | 结论 |
|-----------|---------|------|
| Channel 单例慢 3 倍 | 慢 40–280 倍（取决实现） | 源文章严重低估，方向一致 |
| atomic 比 Mutex 快 3.7 倍 | 无争用仅 1.1x；争用时 4.6x | 只在争用场景成立 |
| WaitGroup 1000 worker 慢 14 倍 | 100→1000 慢 ≈11x（同方向） | 方向一致，量级相近 |
| （未提 RWMutex 纯读） | 纯读下 RWMutex 反而慢 1.6x | 新增反直觉发现 |
| （未提 errgroup 开销） | errgroup 与 WaitGroup 持平 | 换 errgroup 是为能力非速度 |

## 三、复现方式

各实验独立 module，进入 `evidence/code/{实验名}/` 执行：
```bash
GOMAXPROCS=14 go test -bench=. -benchmem -cpu=14 -run=^$ -timeout 240s
```
输出见同目录 `evidence/output/{实验名}/result.txt`。

## 四、自造占比与引用

- 自造论据：8/8 = 100%（无外部引用兜底）
- 外部引用：0 处（唯一可能的"引用"是标准库源码本身，属公共 API，不计入外部引用）
- 降级项：无
