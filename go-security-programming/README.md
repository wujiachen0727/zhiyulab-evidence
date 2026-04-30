# Evidence 总索引

本目录收录《Go 的安全是两层的：一层语言给，一层你自己给》一文的所有自造论据。

所有代码可独立运行，无私有路径依赖。

## 代码论据

| ID | 目录 | 类型 | 核心假设 | 运行方式 |
|:--:|------|:----:|---------|---------|
| E1 | `code/gosec-scan/` | 实验验证 | 真实 Go 项目的高危安全问题主要在第二层 | `./run.sh && python3 aggregate.py` |
| E2 | `code/poc-sql-order-injection/` | 实验验证 | Go 强类型系统不能阻止 ORDER BY 注入 | `go run main.go` |
| E3 | `code/poc-random-misuse/` | 实验验证 | math/rand 生成的 token 不是密码学安全 | `go run main.go` |
| E4 | `code/poc-race-authbypass/` | 实验验证 | goroutine 隔离不保证 TOCTOU 原子性 | `go run -race main.go` |
| E5 | `code/poc-unsafe-pointer/` | 实验验证 | unsafe.Pointer 绕过 Go 类型封装 | `go run main.go` |
| E6 | `code/poc-ssrf-default-client/` | 实验验证 | http.Client 默认信任任何 URL | `go run main.go` |
| E7 | `code/poc-timing-attack/` | 实验验证 | `==` 字符串比较有可测量时序差 | `go test -bench=. -benchtime=3s -count=5` |
| E8 | `code/poc-jwt-alg-none/` | 实验验证 | JWT 手写 verify 不校验 alg 会放行 alg:none | `go run main.go` |

## 数据论据

| ID | 文件 | 描述 |
|:--:|------|------|
| E1 | `data/gosec-distribution.md` | 10 个主流 Go 开源项目 gosec 扫描结果分析报告 |
| E1 | `data/gosec-distribution.json` | 扫描结果聚合 JSON |

## 运行输出归档

| 目录 | 内容 |
|------|------|
| `output/poc-sql-order-injection/result.txt` | E2 运行输出 |
| `output/poc-random-misuse/result.txt` | E3 运行输出 |
| `output/poc-race-authbypass/result.txt` | E4 race detector 输出 |
| `output/poc-unsafe-pointer/result.txt` | E5 运行输出 |
| `output/poc-ssrf-default-client/result.txt` | E6 运行输出 |
| `output/poc-timing-attack/bench-result.txt` | E7 benchmark 结果 |
| `output/poc-jwt-alg-none/result.txt` | E8 运行输出 |

## 环境信息

- **Go**：1.26.2 darwin/arm64
- **CPU**：Apple M4 Pro
- **gosec**：dev @ 2026-04
- **govulncheck**：v1.3.0

## 证伪记录（关键）

**E1 证伪结果**：立意阶段假设"80% 真实漏洞出在第二层"数据不支撑。实际按总数看卫生类（未处理错误）占 67%。但按 HIGH 严重度看，**100%（220/220）的真正高危都在第二层**——这是更精确也更锋利的说法。论点已据此修正。

**E7 证伪结果**：在 Apple M4 Pro + Go 1.26.2 上，`==` 字符串比较的时序差异真实存在（短路生效）：
- 第一字节就不同：~1.40 ns
- 最后一字节才不同：~1.67 ns
- 完全相同：~1.85 ns
`ConstantTimeCompare` 稳定在 ~8.9 ns 无差异。**假设成立。**
