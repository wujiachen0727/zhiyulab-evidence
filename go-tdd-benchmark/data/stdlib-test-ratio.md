# E2: Go 标准库 Test/Bench 比例统计

[实测 Go 1.26.2] 2026-04-14

## 测试方法

统计 Go 标准库（GOROOT/src）中所有 *_test.go 文件的测试函数分布。

环境：Go 1.26.2 darwin/arm64

## 原始数据

| 指标 | 数量 |
|------|------|
| 标准库 *_test.go 文件 | 1753 |
| func TestXxx | 9198 |
| func BenchmarkXxx | 1915 |
| func FuzzXxx | 56 |
| t.Run() 子测试 | 1803 |
| b.Run() 子测试 | 798 |
| t.Helper() | 2559 |
| testify 引用 | 0 |
| 第三方 assert 风格调用 | 0 |
| if 条件判断模式 | 24971 |

## 比例分析

| 类型 | 数量 | 占顶层测试函数比例 |
|------|------|:--:|
| Test | 9198 | 82.3% |
| Benchmark | 1915 | 17.1% |
| Fuzz | 56 | 0.5% |

- Test : Benchmark = 4.8 : 1
- 标准库零 testify / 零第三方 assert
- t.Helper() 2559 处——官方大量自建断言辅助函数

## 结论

1. **Go 标准库不依赖任何第三方 assert 库**——0 处 testify、0 处第三方 assert 风格调用。这验证了"没有 assert 是设计意图"而非遗漏。
2. **Benchmark 占比 17.1%**——不低。Go 标准库对性能验证的重视程度远超一般项目的测试实践。
3. **t.Helper() 使用 2559 次**——Go 官方不是不用断言，而是自建断言辅助函数。这印证了"断言不是框架的职责，是你自己的事"的设计哲学。
4. **if 条件判断模式 24971 处**——标准库确实用 if+Errorf 替代了 assert 库。
