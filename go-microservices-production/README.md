# Evidence 索引 — go-microservices-production

| ID | 类型 | 路径 | 正文引用点 |
|----|------|------|-----------|
| E1 | 实验 | `code/chain-pool/` → `output/chain-pool/result.md` | 三跳链 MaxOpen=5：P99 549ms，WaitCount 353 |
| E4 | 实验 | `code/breaker-trip/` → `output/breaker-trip/result.md` | NotFound 误 trip：open vs closed |
| E5 | 算术 | `code/conn-budget/` → `data/conn-budget.md` | 三服务 275 连接 vs PG 上限 80 |
| E7 | 实验 | `code/pool-churn/` → `output/pool-churn/result.md` | MaxIdleClosed 18 vs 0 |
| E6 | 推演 | `data/timeout-chain.md` | 三层 3s → 最坏 9s |
| E2 | 共识 | `data/otel-memory.md` | ParentBased 10% 基线 |

## 运行方式

```bash
cd evidence/code/pool-churn && go run .
cd evidence/code/breaker-trip && go run .
cd evidence/code/conn-budget && go run .
cd evidence/code/chain-pool && go run .
```

环境：Go 1.24+（breaker-trip 模块 go mod tidy 后需 Go 1.25+ toolchain）
