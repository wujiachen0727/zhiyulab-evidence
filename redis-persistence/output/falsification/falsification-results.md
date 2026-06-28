# 证伪实验结果

**实验时间**：2026-06-28
**环境**：Docker 29.4.0 (Colima) + Redis 7.4.9 (redis:7-alpine)
**Host**：macOS Darwin 25.5.0 ARM64

---

## F1: E1 证伪——极小数据集（10 key）文件大小

**重测说明**：初次实验 AOF 大小为 0 是因路径判断错误。重测发现 Redis 7.4 AOF 采用 multi-part 结构（base.rdb + incr.aof + manifest）。

| 持久化方式 | 文件大小（字节）| 说明 |
|-----------|:-------------:|------|
| RDB (dump.rdb) | 245 | 纯 RDB 二进制 |
| AOF base.rdb | 245 | aof-use-rdb-preamble 默认启用，base file 用 RDB 格式 |
| AOF incr.aof | 0 | 重写后无新写入 |
| AOF manifest | 88 | 清单文件 |
| AOF 总计 | 333 | base + incr + manifest |

**原假设**：RDB 文件比 AOF 文件小
**证伪判断**：⚠️ 部分证伪成立——极小数据集下 AOF 的 base.rdb 与 dump.rdb 完全一致（都是 245 字节），AOF 多出 manifest 开销（88 字节）。纯数据部分无差异
**论点修正**：需修正为"aof-use-rdb-preamble 启用后（7.0+ 默认），AOF 的 base file 与 RDB 文件大小一致；AOF 总大小 = RDB 大小 + 增量日志 + manifest 开销。大数据集下增量日志主导，AOF 显著大于 RDB"
**重要发现**：这验证了"混合持久化"的本质——AOF 重写产出的 base file 就是 RDB 格式，理解了这一点就能理解为什么混合持久化不是"新机制"而是"AOF 重写的默认行为"

## F2: E2 证伪——极小数据集（10 key）恢复时间

**实验状态**：⚠️ 数据无效——RDB 恢复容器启动失败（数据卷挂载问题），AOF 恢复容器启动失败。两者恢复时间数据均为容器启动时间，非 Redis 加载时间

| 持久化方式 | 恢复时间（ms，含容器启动）| 恢复后 key 数 |
|-----------|:-----------------------:|:------------:|
| RDB | 2009 | （容器未运行，未采集）|
| AOF | 2013 | （容器未运行，未采集）|

**原假设**：RDB 恢复比 AOF 快
**证伪判断**：⚠️ 无法判定——极小数据集下恢复时间（< 10ms）被容器启动时间（~2000ms）完全淹没，无法区分
**论点修正**：E2 正式实验必须用大数据集 + Redis 进程内计时（INFO persistence 的 loading_start_time / loading_total_time），消除容器启动噪声。预期：大数据集下 RDB 恢复显著快于 AOF（RDB 是二进制紧凑格式，AOF 需重放命令）

## F3: E3 证伪——空闲无写入场景 BGSAVE 是否触发 RSS 翻倍

**基础数据**：50 万 key，used_memory = 41,287,336 字节（约 39.4MB）

| 采样点 | used_memory_rss（字节）| RSS 变化 |
|--------|:-------------------:|:--------:|
| BGSAVE 前 | 49,283,072（47.00M）| 基准 |
| BGSAVE 中（1s）| 49,414,144（47.12M）| +0.28% |
| BGSAVE 中（2s）| 49,414,144（47.12M）| +0.28% |
| BGSAVE 后 | 49,414,144（47.12M）| +0.28% |
| rdb_last_cow_size | 454,656（约 0.44MB）| — |

**原假设**：fork 期间 RSS 翻倍
**证伪判断**：❌ 证伪成立——空闲无写入场景 RSS 仅增长 0.28%，未翻倍。COW 复制比例 = 454,656 / 41,287,336 ≈ 1.1%，几乎不触发复制
**论点修正**：需修正为"fork 期间 RSS 翻倍是写入密集场景特定现象，空闲场景不翻倍"——这反而验证了 COW 机制的精确性：COW 只在写入时复制，无写入就无复制
**重要意义**：证伪成立验证了机制的精确性。E3 正式实验必须在写入密集场景下测，才能观察到 RSS 翻倍现象

## F4: E7 证伪——fsync 与 fork+COW 独立性

**原假设**：RDB 和 AOF 优缺点都指向 fork+COW 同一根因

### 证伪分析

AOF 的"慢"来自两个独立机制：

1. **fsync 策略**（always / everysec / no）
   - always：每条命令 fsync，吞吐量大幅下降
   - everysec：每秒 fsync 一次，最多丢 1 秒
   - no：由 OS 决定 fsync 时机，性能最好但数据安全性最低
   - **fsync 是磁盘 I/O 操作，与 fork+COW 无关**

2. **fork + COW**（AOF 重写时）
   - AOF 重写触发 fork()，子进程写新 AOF
   - 重写期间父进程新命令缓存在 aof_rewrite_buffer
   - 重写完成后替换旧 AOF

3. **RDB 的"慢"只来自 fork + COW**
   - RDB 没有 fsync 策略选择，BGSAVE 就是 fork + COW
   - RDB 的数据丢失窗口 = save 规则间隔，与 fsync 无关

### 证伪结论

**❌ 证伪成立**——AOF 的性能/安全权衡由 fsync 策略和 fork+COW 两个独立机制共同决定，不能简单归因为 fork+COW 单一根因。

### 论点修正

原论点："三种持久化机制的优缺点，本质都指向同一个底层机制 fork()+COW"

修正为："三种持久化机制的优缺点，**主要根因**都指向 fork()+COW——RDB 完全靠它，AOF 重写靠它，混合持久化的 base file 生成也靠它。但 AOF 多了一个独立的 fsync 策略维度，这是 fork+COW 解释不了的。"

### 修正后的完整论点链

- RDB 优缺点 → fork+COW（唯一根因）
- AOF 优缺点 → fork+COW（重写时）+ fsync 策略（日常写入时）——**双根因**
- 混合持久化优缺点 → fork+COW（AOF 重写时用 RDB 做 base file）+ fsync 策略（增量部分）

**fsync 策略是 AOF 独有的维度，不能用 fork+COW 解释。** 这是 E7 证伪最重要的修���——让论点更精确，避免被冷读视角4（魔鬼代言）攻击。

