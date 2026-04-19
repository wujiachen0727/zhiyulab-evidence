# E2：等待陷阱证伪实验

## 证伪假设

**原假设（立意阶段）**：pprof CPU profile 无法识别"等待导致的慢"。

**证伪结果**：**精确化成立**。原假设过于粗糙，实测后调整为：

> pprof 能告诉你**累计等了多久**，但 trace 才能告诉你**什么时候等、等了多少次、跟谁有关**。

## 实验目录结构

```
e2-wait-trap/
├── README.md                         ← 本文件
├── run1-original/                    ← 第一次失败的实验（保留作为教训）
│   ├── main.go                        （channel 主导，锁争用没起来）
│   └── output/                        （失败的 profile 数据）
└── e2a-mutex-contention/             ← 成功的证伪实验
    ├── main.go
    └── output/
        ├── cpu.pprof
        ├── mutex.pprof
        ├── block.pprof
        └── trace.out
```

## 最终的双实验叙事

**Run1（失败）→ E2a（成功）**的演进本身就是文章素材：

1. **Run1 教训**：想证明"锁争用在 pprof 里不可见"时，给 worker 加了 channel 阻塞下游——结果 channel 成了"天然节流阀"，锁争用根本没起来，mutex profile 为空。
2. **重做 E2a**：移除 channel 干扰，延长临界区到 6-10μs 级别，50 个 worker 对 1 把锁的重度争用终于出现。

这个"第一次实验翻车 → 数据告诉我为什么 → 重做才看到真相"的叙事，在"锐利观点"风格下有两个价值：
- 展示实验思维（不是一次性跑成功）
- 自动回应"你的数据可信吗"的质疑——过程可追溯

## 运行环境

- Go 1.26.2
- darwin/arm64（Apple Silicon）
- GOMAXPROCS=14

## 关键数据（E2a，详见 output/e2-wait-trap/e2a-summary.md）

| 维度 | CPU Profile | Mutex Profile | Block Profile | Trace |
|------|:----------:|:-------------:|:-------------:|:-----:|
| 能看见锁等待总量 | ❌ | ✅ 64s | ✅ 67s | ✅ 84s |
| 锁等待归因点 | N/A | **Unlock**（反直觉） | Lock | 每次具体事件 |
| 能看见"抢不到 CPU"的等待 | ❌ | ❌ | ❌ | ✅ 0.99s |
| 能看见单 goroutine 分布 | ❌ | ❌ | ❌ | ✅ |
| 能看见阻塞次数 | ❌ | ❌ | ❌ | ✅ 平均 2000+ |
| 文件大小 | 1.2 KB | 1.9 KB | 1.5 KB | 3.6 MB |

## 复现命令

```bash
# 失败版（Run1）
cd run1-original && mkdir -p output && go run main.go

# 成功版（E2a）
cd e2a-mutex-contention && mkdir -p output && go run main.go

# 查看 profile
go tool pprof -top output/cpu.pprof
go tool pprof -top output/mutex.pprof
go tool pprof -top output/block.pprof
go tool trace output/trace.out  # 浏览器打开
```
