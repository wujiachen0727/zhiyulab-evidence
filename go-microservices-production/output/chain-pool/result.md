# 三跳链 + 小连接池实验

环境：go1.24.13；MaxOpen=5，并发=30，每请求 3 次 DB 查询

| P99 延迟 | WaitCount | WaitDuration | InUse 峰值观察 |
|---|---:|---:|---:|
| 549.7ms | 353 | 25.99s | 0 |
