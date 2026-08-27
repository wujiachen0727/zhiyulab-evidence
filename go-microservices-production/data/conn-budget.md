# 连接预算算术实验

假设 PostgreSQL max_connections=100，预留 headroom=80% → 可用上限 80

| 服务 | 副本 | MaxOpen | 小计 | 累计 | 状态 |
|---|---:|---:|---:|---:|---|
| order | 4 | 25 | 100 | 100 | OVER |
| inventory | 4 | 25 | 100 | 200 | OVER |
| payment | 3 | 25 | 75 | 275 | OVER |

结论：三服务合计峰值连接 275，超过 PG 可用上限 80 共 195。
