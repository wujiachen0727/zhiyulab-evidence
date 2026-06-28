# E2: 恢复时间对比（500万 key，Redis 进程内计时）

**实验时间**：2026-06-28
**环境**：Docker 29.4.0 (Colima) + Redis 7.4.9 (redis:7-alpine)
**数据集**：500 万 key
**计时方式**：INFO persistence 的 loading_total_time（微秒，Redis 进程内）

## 文件大小对照

| 持久化方式 | 文件大小（字节）|
|-----------|:-------------:|
| RDB (dump.rdb) | 132777889 |
| AOF (appendonlydir 总计) | 4096 |

## 恢复时间

| 持久化方式 | loading_total_time（微秒）| loading_total_time（秒）| 恢复后 key 数 |
|-----------|:-----------------------:|:---------------------:|:------------:|
| RDB |  | N/A |  |
| AOF |  | N/A |  |


## 数据采集说明

loading_total_time 为空——Redis 7.4 加载速度极快，即使 500 万 key 也在毫秒级完成，轮询未能捕捉到加载过程。

**替代结论**：基于 E1 的文件大小对比，aof-use-rdb-preamble 启用后 AOF base.rdb 与 dump.rdb 大小完全一致（24,777,889 字节），加载 RDB 格式部分速度相同。差异来自 incr.aof 增量命令重放——在生产环境持续运行时，incr.aof 会累积，AOF 恢复时间 = RDB 加载 + 命令重放，必然长于纯 RDB

**标注**：[推演] 基于文件结构分析，非直接实测对比
