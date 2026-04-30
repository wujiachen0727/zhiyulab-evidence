# E7: 时序攻击 benchmark — `==` vs ConstantTimeCompare

## 证伪结果

**假设**：`==` 字符串比较有可测量的时序差异（早退）。  
**结果**：**假设成立**（在 Apple M4 Pro + Go 1.26.2 上）。

## Benchmark 数据（5 次运行平均）

| 比较方式 | 不匹配位置 | ns/op | 差异 |
|---------|-----------|:-----:|:----:|
| `==` | 第 1 字节 | ~1.40 | 基准 |
| `==` | 最后字节 | ~1.67 | +19% |
| `==` | 完全相等 | ~1.85 | +32% |
| `subtle.ConstantTimeCompare` | 任意位置 | ~8.9 | 恒定 |

**关键观察**：`==` 比较在越早发现差异时越快，差异最大处有 35% 的时序差异（0.45 ns）。
`ConstantTimeCompare` 无论输入如何都稳定在 ~8.9 ns。

## 防 DCE 优化

Benchmark 中使用 package-level `sinkBool` 变量保存返回值——
避免编译器把整个比较调用优化掉（来自 go-network-programming 复盘教训）。

## 运行

```bash
go test -bench=. -benchmem -count=5 -benchtime=3s
```

## 对应章节

第三章 第 6 个幻觉 · 比较字符串用 == 就行

## 威胁模型说明

单次纳秒级差异在远程网络攻击中较难利用（网络抖动 > 时序差）。
但在本地/同机/侧信道攻击场景下，大量统计后可恢复——这是 crypto/subtle 存在的理由。
"业务用 == 是生产隐患"不是夸张，是防御深度原则。

## 环境

- Go 1.26.2 darwin/arm64
- CPU：Apple M4 Pro（ARM）
- 不同 CPU 和 Go 版本数值会变，但**趋势一致**（越晚不匹配越慢）
