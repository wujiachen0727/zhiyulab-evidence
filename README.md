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
| [《Redis 过期策略 vs 内存淘汰：你分清了吗？》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/redis-expiry-eviction) | 2026-06-26 | `redis-expiry-eviction/` | 4 组实验 + 1 组场景（E1 惰性删除 + E2 定期删除 + E3 淘汰策略 + E4 LRU samples + E5 session 雪崩场景） |
| [《Go 协程池：高并发场景下的资源管控必修课》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-goroutine-pool) | 2026-06-24 | `go-goroutine-pool/` | 3 组（benchmark + OOM demo + 流量突增） |
| [《数据所有权：预防 Go data race 的设计思维》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-data-race-beyond) | 2026-06-19 | `go-data-race-beyond/` | 4 组（data-race-demo、ownership-model-demo、ownership-transfer-demo、pipeline-demo） |
| [《Go HTTP 请求慢在哪里？》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-httptrace-tracing) | 2026-06-14 | `go-httptrace-tracing/` | 3 组（httptrace CLI + 连接复用对比 + 多站点阶段耗时） |
| [《三行代码就能卡住你的 Go 服务——不可见的并发阻塞模式》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-deadlock-unreported) | 2026-06-13 | `go-deadlock-unreported/` | 5 组（chan 阻塞 + context 断链 + rwmutex 递归读 + 调用链 + 信号参考） |
| [《WebSocket 是个好东西，但你不需要它——从 AI 流式到实时推送，SSE 的逆袭》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/sse-vs-websocket) | 2026-06-12 | `sse-vs-websocket/` | 1 组（benchmark） |
| [《设计模式是 Go 的第二语言》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-design-pattern-practice) | 2026-06-09 | `go-design-pattern-practice/` | 5 组（e1-e5） |
| [《冷启动雪崩的三种策略：惰性加载、主动预热、渐进式预热怎么选》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/cache-warmup-strategies) | 2026-06-09 | `cache-warmup-strategies/` | 1 组（缓存策略模拟） |
| [《Go 反射的暗债：encoding/json 为什么不用代码生成》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-json-reflection-debt) | 2026-06-02 | `go-json-reflection-debt/` | 1 组（v1/v2/jsoniter/sonic 四库 benchmark + K8s API struct 统计） |
| [《限流：令牌桶、漏桶、滑动窗口怎么选》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/rate-limiter-algorithms) | 2026-06-02 | `rate-limiter-algorithms/` | 4 组（滑动窗口三变体 + 漏桶 vs Nginx + 分布式令牌桶 + ZSET 内存） |
| [《Claude Code 用着用着就忘——是它的上下文机制，不是它的记忆力》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/claude-code-context-management) | 2026-06-01 | `claude-code-context-management/` | 1 组（三级压缩阈值工程意义验证 + 多场景触发轮次估算） |
| [《分布式锁不是选 Redis 就完事了》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/distributed-lock-selection) | 2026-05-31 | `distributed-lock-selection/` | 6 组（E1-E5 实验代码 + E6 benchmark） |
| [《「诚实」是新的「聪明」——Claude 4.8 对 AI 评价体系的三重追问》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/claude-opus-48-release) | 2026-05-31 | `claude-opus-48-release/` | 1 组（诚实度对比实验） |
| [《为什么大厂还在用 RPC？不是因为快，是因为不崩》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/rpc-vs-http) | 2026-05-30 | `rpc-vs-http/` | 2 组（协议层吞吐对比 E1 + 端到端延迟分解 E2） |
| [《从 PHP 到 Go：真正迁移的是复杂度的归属》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/php-to-go-migration) | 2026-05-30 | `php-to-go-migration/` | 1 组对照实验（PHP weak / PHP strict / Go decode） |
| [《泛型的本质，是把混乱挡在编译期门口》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/generics-convergence) | 2026-05-30 | `generics-convergence/` | 1 组四路径对照实验 |
| [《并发模型三流派：CSP / Actor / 线程》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/concurrency-models) | 2026-05-30 | `concurrency-models/` | 8 组（Go/Java/Erlang 三模型任务编排器对照 + 场景推演） |
| [《为什么 Python 的简单越到工程里越贵？》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/python-simplicity-cost) | 2026-05-30 | `python-simplicity-cost/` | 6 组（类型迟绑定 + GIL + 脚本到项目成本 + 三层成本框架 + 交付清单） |
| [《包管理器不是下载器，是构建信任的三层协议》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/package-manager-evolution) | 2026-05-30 | `package-manager-evolution/` | 1 组（Go/Npm/Cargo/Python 依赖文件解剖） |
| [《Go map 的不安全，其实是一条数据可信度红线》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-map-fatal-not-panic) | 2026-05-29 | `go-map-fatal-not-panic/` | 2 组（recover vs fatal + map sync benchmark） |
| [《别再背 slice 扩容公式了：1.18 真正改掉了什么》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/slice-growth-go118) | 2026-05-28 | `slice-growth-go118/` | 5 组自造证据 + 1 组官方快照 |
| [《好的 DX 不等于少写代码——三种语言的摩擦力设计课》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/language-design-friction) | 2026-05-28 | `language-design-friction/` | 5 组 |
| [《一次接口超时排查：从应用层挖到 TCP 内核参数》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/tcp-keepalive-nat-timeout) | 2026-05-28 | `tcp-keepalive-nat-timeout/` | 1 组（keepalive 矩阵分析） |
| [《从 Vibe Coding 到可交付工程，中间差一套刹车系统》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/vibe-coding-serious-engineering) | 2026-05-28 | `vibe-coding-serious-engineering/` | 1 组（交付差距分析） |
| [《DDD 落地：你的团队扛得住七层间接吗？》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/ddd-three-conditions) | 2026-05-26 | `ddd-three-conditions/` | 2 组 |
| [《Go 泛型两年后：反射可以退休了吗》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-generics-vs-reflection) | 2026-05-21 | `go-generics-vs-reflection/` | 5 组 |
| [《缓存穿透/击穿/雪崩：面试能背，上线能用吗》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/cache-breakdown-myths) | 2026-05-16 | `cache-breakdown-myths/` | 3 组（布隆过滤器内存 + 锁竞争 benchmark + 预热 demo） |
| [《一次 goroutine 泄漏：pprof 说有 10 万个 goroutine，但问题不在 channel》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/goroutine-leak-triple-combo) | 2026-05-15 | `goroutine-leak-triple-combo/` | 2 组（泄漏复现 + 修复） |
| [《sync.Pool 的真正分界线不是对象大小——一次 benchmark 翻车记录》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-sync-pool-pitfall) | 2026-05-14 | `go-sync-pool-pitfall/` | 1 组 |
| [《消息队列是解耦神器还是复杂度放大器》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/do-you-need-mq) | 2026-05-13 | `do-you-need-mq/` | 2 组（outbox vs MQ + 幂等消费） |
| [《为什么所有 AI 工具都在用 TypeScript》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/why-ai-tools-use-typescript) | 2026-05-13 | `why-ai-tools-use-typescript/` | 1 组（zod-to-tooluse demo） |
| [《Go context 超时传播：你以为设了就安全了》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-context-timeout-propagation) | 2026-05-10 | `go-context-timeout-propagation/` | 5 组（计时起点 + 子 context + gRPC vs HTTP + DB 连接池冲突 + 绝对 deadline） |
| [《Go 代码生成的三层认知：从忍住不用到自己造轮子》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-code-generation) | 2026-05-05 | `go-code-generation/` | 2 组 |
| [《你的 SQL 没慢，慢的是 Go 连接池里的队伍》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-database-connection-pool) | 2026-05-02 | `go-database-connection-pool/` | 1 组 |
| [《从文件到配置中心：Go 配置管理的三个升级拐点》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-config-management) | 2026-05-02 | `go-config-management/` | 1 组（stdlib/viper/koanf 三库对照） |
| [《Go 跨平台编译的决策树：从「能编译」到「能部署」的 5 个关键抉择》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-cross-compilation) | 2026-05-02 | `go-cross-compilation/` | 2 组（CGO 对比 + embed 对比） |
| [《Go 的安全是两层的：一层语言给，一层你自己给》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-security-programming) | 2026-04-30 | `go-security-programming/` | 8 组 PoC + 1 组聚合扫描 |
| [《从 sync.Map 到 Redis：Go 缓存升级的三个拐点》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-cache-system) | 2026-04-30 | `go-cache-system/` | 3 组（benchmark + GC 压力 + Redis 延迟） |
| [《Go 日志性能：5 个设计决策，比选库重要得多》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-logging-design) | 2026-04-30 | `go-logging-design/` | 1 组 |
| [《写 Go linter 不难，难的是让团队用起来》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-custom-linter) | 2026-04-26 | `go-custom-linter/` | 3 组 |
| [《别只会写 net.Listen：Go 网络编程的三层进阶》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-network-programming) | 2026-04-26 | `go-network-programming/` | 4 组 |
| [《为什么你的 Go TCP server P99 延迟这么高》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-tcp-diagnostics) | 2026-04-26 | `go-tcp-diagnostics/` | 3 组 |
| [《别用 Go 写插件系统——但如果你非要写，这里有张决策表》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-plugin-system) | 2026-04-20 | `go-plugin-system/` | 2 组（benchmark + reload） |
| [《从 pprof 到持续 profiling：Go 性能工具链的三次升级》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-profiling-toolchain) | 2026-04-19 | `go-profiling-toolchain/` | 4 组（8 条独立论据） |
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
| [《Go GC 十年：一部延迟战争史》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-gc-deep-dive) | 2026-04-08 | `go-gc-deep-dive/` | 4 组（GC trace + GOGC 对比 + GOMEMLIMIT + 碎片化） |
| [《Claude Agent Teams 实战手册：从零开始搭建多 Agent 系统》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/claude-agent-teams) | 2026-03-03 | `claude-agent-teams/` | 1 组（单/多 Agent 系统对比测试） |

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
