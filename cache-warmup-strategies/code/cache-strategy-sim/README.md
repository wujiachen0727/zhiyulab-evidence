# Cache Strategy Simulator

模拟三种缓存冷启动策略在相同条件下的行为差异。

## 环境要求

- Go 1.21+
- 无外部依赖

## 运行

```bash
cd evidence/code/cache-strategy-sim
go run main.go
```

## 模拟参数

| 参数 | 值 | 说明 |
|------|:---:|------|
| TotalKeys | 10,000 | 缓存总键数 |
| HotKeyRatio | 20% | 20% 的键占 80% 的访问 |
| Concurrency | 50 | 并发 worker 数 |
| RequestCount | 10,000 | 总请求数 |
| WarmupBatchSize | 200 | 渐进式预热每批大小 |
| WarmupInterval | 2ms | 渐进式预热每批间隔 |

## 三种策略

1. **惰性加载+保护**：请求时才加载，超过 50 并发 DB 查询触发保护
2. **主动预热**：启动时加载全部热点键
3. **渐进式预热**：分批加载热点键，每批 200 个，间隔 2ms，未命中时限流 30
