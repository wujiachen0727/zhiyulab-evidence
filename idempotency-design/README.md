# Evidence 总索引

> 文章：幂等性设计：唯一ID、状态机与乐观锁的三层防线
> 环境：Go 1.26.2 (darwin/arm64)，全部实验纯标准库、无外部依赖，本地可复现

## 论据结构

| ID | 类型 | 内容 | 代码 | 输出 | 状态 |
|----|------|------|------|------|:----:|
| E1 | 实验验证 | 幂等键防重复：同请求重发 100 次，无幂等键 99 笔重复入账，有幂等键 0 笔 | `code/e1-idempotency-key/` | `output/e1-idempotency-key/result.txt` | ✅ |
| E2 | 实验验证 | 状态机防乱序：乱序回调被拒，重复回调幂等 | `code/e2-state-machine/` | `output/e2-state-machine/result.txt` | ✅ |
| E3 | 实验验证 | 乐观锁防并发：10 笔并发扣款无锁丢 9 笔，乐观锁全生效（冲突重试 45 次） | `code/e3-optimistic-lock/` | `output/e3-optimistic-lock/result.txt` | ✅ |
| E4 | 实验验证 | 层间冲突：盲目重试=渠道退款 2 次（重复退款），幂等键拦截=1 次 | `code/e4-layer-conflict/` | `output/e4-layer-conflict/result.txt` | ✅ |

## 外部引用（≤3 处）

| ID | 内容 | 来源 |
|----|------|------|
| R1 | Stripe 幂等键机制（保存首次结果、参数校验、24h 清理） | Stripe 官方文档 |
| R2 | Google Cloud 幂等实现流程 + HTTP 方法幂等性（RFC 9110） | Google Cloud 文档 |

## 运行方式

```bash
# 每个实验独立运行
cd evidence/code/e{1..4}-{name}/ && go run main.go
```

## 推演/实测标注

- E1-E4 全部为 [实测 Go 1.26.2]（本地运行结果）
- 并发实验（E1/E3/E4）含时序依赖，冲突次数可能小幅波动，但结论方向稳定
