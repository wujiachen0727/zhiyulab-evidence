# 连接池 churn 对比实验

环境：go1.24.13；并发 40，请求 200，单次查询 sleep 20ms

| 配置 | P99 | MaxIdleClosed | WaitCount |
|---|---:|---:|---:|
| churn: MaxIdle=2, MaxOpen=20 | 83.2ms | 18 | 180 |
| warm: MaxIdle=20, MaxOpen=20 | 102.7ms | 0 | 180 |
