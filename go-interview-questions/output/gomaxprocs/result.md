# gomaxprocs 运行结果

**对应题目**：十、`GOMAXPROCS`
**目的**：确认 `GOMAXPROCS` 管的是什么、调大是否一定更快、容器里的默认值
**运行方式**：`cd code/gomaxprocs && go run .`
**原始输出**：`result.txt`（stdout）

## 测试方法

4 个 CPU 密集协程，每个做 1.5 亿次累加。每一档先热身一次，再取 3 次运行的最小值——单次计时会被调度抖动和机器负载带偏，取最小值更接近这一档的真实下界。

环境：`runtime.NumCPU() = 14`。

## 关键输出

```
runtime.NumCPU()       = 14
runtime.GOMAXPROCS(0)  = 14

GOMAXPROCS=1 -> 139ms    相对 1 的加速比 1.00x
GOMAXPROCS=2 -> 70ms     相对 1 的加速比 1.98x
GOMAXPROCS=4 -> 38ms     相对 1 的加速比 3.65x
GOMAXPROCS=8 -> 38ms     相对 1 的加速比 3.67x
```

## 结论

第一，`GOMAXPROCS` 管的不是协程数量。协程你还是可以开几十万个，它决定的是同一时刻有几个线程在真正跑 Go 代码。

第二，加速比在协程数用尽处停住。从 1 到 4 提速 3.65 倍，接近线性；从 4 到 8 几乎没动，只多出 0.02x，因为可跑的协程只有 4 个。

第三，跟容器有关：从 Go 1.25 起，在 Linux 上默认值会考虑容器 cgroup 的 CPU 带宽限制（Kubernetes 的 CPU limit，不含 CPU request），并在限制变化时定期更新。

**复现注意**：绝对值会随机器负载浮动（实测波动约 ±3%），但"在协程数用尽处停止增长"这一形状稳定复现。
