# zhiyulab-evidence

> 止语Lab 技术文章配套的实验代码和实测数据——可信度承诺的底稿。

## 为什么有这个仓库

止语Lab 的长文都追求"论据自造"：能自己跑实验就不引用别人的结论。每一个实测数字、每一张 profile 都来自我在本地搭环境跑出来的原始文件。

这个仓库把这些实验代码和实测数据全部公开，目的有三：

1. **可复现**：读者可以亲自跑一遍，验证文章里的数字
2. **可质疑**：数据来源透明，谁都能翻看原始 pprof / trace / log 文件
3. **可继承**：实验代码本身可能比文章有更长的生命力——后来人可以在此基础上做自己的实验

## 目录结构

按文章 slug 分子目录，每篇文章一个独立空间。

## 文章清单

> 按发布时间倒序。点击文章标题跳转到对应的子目录。

| 文章 | 发布时间 | 子目录 | 配套实验数 |
|------|---------|--------|:---------:|
| [《别用 Go 写插件系统——但如果你非要写，这里有张决策表》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-plugin-system) | 2026-04-20 | `go-plugin-system/` | 2 组（benchmark + reload）|
| [《从 pprof 到持续 profiling：Go 性能工具链的三次升级》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-profiling-toolchain) | 2026-04-19 | `go-profiling-toolchain/` | 4 组（8 条独立论据）|
| [《Gin 很好，但你的项目可能需要更多》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-web-framework-design) | 2026-04-16 | `go-web-framework-design/` | 3 组 |
| [《别急着拆微服务：Go 项目演进的三个关键决策》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-microservice-patterns) | 2026-04-16 | `go-microservice-patterns/` | 2 组 |
| [《从手动到框架：Go DI 演进的三个拐点》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-di-evolution) | 2026-04-15 | `go-di-evolution/` | 5 组 |
| [《Go vs Java GC：同一场延迟战争的两条路》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-vs-java-gc) | 2026-04-15 | `go-vs-java-gc/` | 1 组 |
| [《Go 的测试框架不想让你 TDD》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-tdd-benchmark) | 2026-04-14 | `go-tdd-benchmark/` | 3 组 |
| [《你写的Go代码，编译器真的看得懂吗》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-compiler-optimization) | 2026-04-13 | `go-compiler-optimization/` | 5 组 |
| [《Go 反射为什么"难用"？因为它本来就不想让你用》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-reflect-deep-dive) | 2026-04-12 | `go-reflect-deep-dive/` | 2 组 |
| [《Go 错误分层实战：从裸奔到三层防线》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-error-handling) | 2026-04-11 | `go-error-handling/` | 3 组 |
| [《从一行超时配置到分布式可观测性——Go HTTP 服务的渐进式演进实战》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-http-optimization) | 2026-04-10 | `go-http-optimization/` | 6 组 |
| [《Go 并发编程实战：Channel 还是 Mutex？一个场景驱动的选择框架》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-channel-vs-mutex) | 2026-04-09 | `go-channel-vs-mutex/` | 2 组 |
| [《Go 内存管理优化：内联是逃逸分析的隐藏杠杆》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-escape-analysis) | 2026-04-09 | `go-escape-analysis/` | 3 组 |

（后续文章发布时会在此追加）

## 复现原则

- **二进制不入库**：所有 `.go` / `.java` 源码会入库，但编译产物（可执行文件、`.class`）不入库。跑实验前请自己 `go build` 或 `javac`
- **原始数据入库**：`.pprof`、`.trace`、`.csv`、`.log` 这类实测原始文件会保留
- **子目录名自解释**：大多数子目录按场景命名（如 `inline-vs-escape/`、`timeout-config/`），配合对应文章阅读即可理解意图
- **部分目录有 README**：近期文章（如 `go-profiling-toolchain/`）每个子目录都有独立 README；历史文章可能没有 README，但源码结构直接对应文章内的实验标号

## 联系作者

- 作者：吴嘉晨 / 止语Lab
- 微信公众号：止语Lab
- 反馈：在对应文章下留言，或提 Issue

## License

MIT License（见 `LICENSE` 文件）
