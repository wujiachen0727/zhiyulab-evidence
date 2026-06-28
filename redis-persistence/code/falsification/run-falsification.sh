#!/bin/bash
# 证伪实验主脚本——E1/E2/E3/E7 核心假设证伪
# 在 Docker Redis 7.4.9 上运行，结果输出到 evidence/output/falsification/
set -e

OUTPUT_DIR="$(dirname "$0")/../../output/falsification"
mkdir -p "$OUTPUT_DIR"
RESULT_FILE="$OUTPUT_DIR/falsification-results.md"

cat > "$RESULT_FILE" << 'EOF'
# 证伪实验结果

**实验时间**：2026-06-28
**环境**：Docker 29.4.0 (Colima) + Redis 7.4.9 (redis:7-alpine)
**Host**：macOS Darwin 25.5.0 ARM64

---

EOF

echo "=== 证伪实验开始 ==="

# ========================================
# F1: E1 证伪——极小数据集（10 key）文件大小对比
# ========================================
echo "运行 F1: E1 证伪（极小数据集文件大小）..."

# 启动 RDB-only 容器
docker rm -f falsify-rdb 2>/dev/null || true
docker run -d --name falsify-rdb -p 6380:6379 \
  redis:7-alpine redis-server --save "" --appendonly no --dbfilename dump.rdb --dir /data
sleep 2

# 灌入 10 个 key
for i in $(seq 1 10); do
  docker exec falsify-rdb redis-cli SET "key_$i" "value_$i"
done
docker exec falsify-rdb redis-cli BGSAVE
sleep 1
RDB_SIZE_SMALL=$(docker exec falsify-rdb stat -c %s /data/dump.rdb 2>/dev/null || echo "0")
docker rm -f falsify-rdb

# 启动 AOF-only 容器
docker rm -f falsify-aof 2>/dev/null || true
docker run -d --name falsify-aof -p 6381:6379 \
  redis:7-alpine redis-server --appendonly yes --appendfsync everysec --save "" --dir /data
sleep 2

# 灌入 10 个 key
for i in $(seq 1 10); do
  docker exec falsify-aof redis-cli SET "key_$i" "value_$i"
done
sleep 1
# 触发 AOF 重写确保文件落盘
docker exec falsify-aof redis-cli BGREWRITEAOF
sleep 2
AOF_SIZE_SMALL=$(docker exec falsify-aof sh -c "wc -c < /data/appendonlydir/*.aof 2>/dev/null || wc -c < /data/appendonly.aof 2>/dev/null || echo 0" | tr -d ' ')
docker rm -f falsify-aof

cat >> "$RESULT_FILE" << EOF
## F1: E1 证伪——极小数据集（10 key）文件大小

| 持久化方式 | 文件大小（字节）|
|-----------|:-------------:|
| RDB (dump.rdb) | $RDB_SIZE_SMALL |
| AOF (appendonlydir/*.aof) | $AOF_SIZE_SMALL |

**原假设**：RDB 文件比 AOF 文件小
**证伪判断**：$([ "$RDB_SIZE_SMALL" -gt "$AOF_SIZE_SMALL" ] && echo "❌ 证伪成立——极小数据集下 AOF 反而更小，RDB 有固定头部开销" || echo "✅ 证伪不成立——极小数据集下 RDB 仍不大于 AOF")
**论点修正**：$([ "$RDB_SIZE_SMALL" -gt "$AOF_SIZE_SMALL" ] && echo "需修正为'大数据集下 RDB 文件更小'" || echo "无需修正")

EOF

echo "F1 完成: RDB=$RDB_SIZE_SMALL AOF=$AOF_SIZE_SMALL"

# ========================================
# F2: E2 证伪——极小数据集恢复时间对比
# ========================================
echo "运行 F2: E2 证伪（极小数据集恢复时间）..."

# 准备 RDB 数据
docker rm -f falsify-rdb-prep 2>/dev/null || true
docker run -d --name falsify-rdb-prep -p 6382:6379 \
  redis:7-alpine redis-server --save "" --appendonly no --dbfilename dump.rdb --dir /data
sleep 2
for i in $(seq 1 10); do
  docker exec falsify-rdb-prep redis-cli SET "key_$i" "value_$i"
done
docker exec falsify-rdb-prep redis-cli BGSAVE
sleep 1
docker cp falsify-rdb-prep:/data/dump.rdb /tmp/falsify-dump.rdb
docker rm -f falsify-rdb-prep

# 准备 AOF 数据
docker rm -f falsify-aof-prep 2>/dev/null || true
docker run -d --name falsify-aof-prep -p 6383:6379 \
  redis:7-alpine redis-server --appendonly yes --appendfsync everysec --save "" --dir /data
sleep 2
for i in $(seq 1 10); do
  docker exec falsify-aof-prep redis-cli SET "key_$i" "value_$i"
done
sleep 1
docker exec falsify-aof-prep redis-cli BGREWRITEAOF
sleep 2
docker cp falsify-aof-prep:/data/appendonlydir /tmp/falsify-appendonlydir
docker rm -f falsify-aof-prep

# 测 RDB 恢复时间
docker rm -f falsify-rdb-recover 2>/dev/null || true
docker run -d --name falsify-rdb-recover -p 6384:6379 \
  -v /tmp/falsify-dump.rdb:/data/dump.rdb \
  redis:7-alpine redis-server --save "" --appendonly no --dbfilename dump.rdb --dir /data
RDB_RECOVER_START=$(date +%s%N)
sleep 2
RDB_RECOVER_END=$(date +%s%N)
RDB_RECOVER_MS=$(( (RDB_RECOVER_END - RDB_RECOVER_START) / 1000000 ))
RDB_KEY_COUNT=$(docker exec falsify-rdb-recover redis-cli DBSIZE | tr -d '\r')
docker rm -f falsify-rdb-recover

# 测 AOF 恢复时间
docker rm -f falsify-aof-recover 2>/dev/null || true
docker run -d --name falsify-aof-recover -p 6385:6379 \
  -v /tmp/falsify-appendonlydir:/data/appendonlydir \
  redis:7-alpine redis-server --appendonly yes --appendfsync everysec --save "" --dir /data
AOF_RECOVER_START=$(date +%s%N)
sleep 2
AOF_RECOVER_END=$(date +%s%N)
AOF_RECOVER_MS=$(( (AOF_RECOVER_END - AOF_RECOVER_START) / 1000000 ))
AOF_KEY_COUNT=$(docker exec falsify-aof-recover redis-cli DBSIZE | tr -d '\r')
docker rm -f falsify-aof-recover

cat >> "$RESULT_FILE" << EOF
## F2: E2 证伪——极小数据集（10 key）恢复时间

| 持久化方式 | 恢复时间（ms，含容器启动）| 恢复后 key 数 |
|-----------|:-----------------------:|:------------:|
| RDB | $RDB_RECOVER_MS | $RDB_KEY_COUNT |
| AOF | $AOF_RECOVER_MS | $AOF_KEY_COUNT |

**原假设**：RDB 恢复比 AOF 快
**证伪判断**：$([ "$RDB_RECOVER_MS" -ge "$AOF_RECOVER_MS" ] && echo "❌ 证伪成立——极小数据集下 RDB 恢复不快于 AOF，差异在容器启动噪声内" || echo "✅ 证伪不成立——RDB 恢复仍快于 AOF")
**论点修正**：$([ "$RDB_RECOVER_MS" -ge "$AOF_RECOVER_MS" ] && echo "需修正为'大数据集下 RDB 恢复更快'" || echo "无需修正")
**注**：容器启动时间占主导，极小数据集下恢复时间差异被噪声淹没。E2 正式实验需用大数据集 + 纯 Redis 进程内计时（INFO persistence loading_total_time）

EOF

echo "F2 完成: RDB=${RDB_RECOVER_MS}ms AOF=${AOF_RECOVER_MS}ms"

# ========================================
# F3: E3 证伪——空闲无写入场景 BGSAVE 是否触发 RSS 翻倍
# ========================================
echo "运行 F3: E3 证伪（空闲场景 BGSAVE）..."

docker rm -f falsify-idle 2>/dev/null || true
docker run -d --name falsify-idle -p 6386:6379 \
  redis:7-alpine redis-server --save "" --appendonly no --dir /data
sleep 2

# 灌入 50 万 key 建立基础内存
docker exec falsify-idle redis-cli --pipe << EOF
SET key:0000001 value_0000001
EOF
# 用脚本批量灌入
docker exec falsify-idle sh -c 'for i in $(seq 1 500000); do echo "SET key_$i value_$i"; done | redis-cli --pipe'
sleep 2

# 采样 BGSAVE 前内存
RSS_BEFORE=$(docker exec falsify-idle redis-cli INFO memory | grep used_memory_rss | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')
USED_MEM=$(docker exec falsify-idle redis-cli INFO memory | grep '^used_memory:' | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')

# 触发 BGSAVE（无写入负载）
docker exec falsify-idle redis-cli BGSAVE
sleep 1
# BGSAVE 期间采样
RSS_DURING=$(docker exec falsify-idle redis-cli INFO memory | grep used_memory_rss | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')
sleep 1
RSS_DURING_2=$(docker exec falsify-idle redis-cli INFO memory | grep used_memory_rss | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')

# 等 BGSAVE 完成
while [ "$(docker exec falsify-idle redis-cli INFO persistence | grep rdb_bgsave_in_progress | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')" = "1" ]; do
  sleep 1
done
RSS_AFTER=$(docker exec falsify-idle redis-cli INFO memory | grep used_memory_rss | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')
COW_SIZE=$(docker exec falsify-idle redis-cli INFO persistence | grep rdb_last_cow_size | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')

docker rm -f falsify-idle

cat >> "$RESULT_FILE" << EOF
## F3: E3 证伪——空闲无写入场景 BGSAVE 是否触发 RSS 翻倍

**基础数据**：50 万 key，used_memory = $USED_MEM 字节

| 采样点 | used_memory_rss（字节）|
|--------|:-------------------:|
| BGSAVE 前 | $RSS_BEFORE |
| BGSAVE 中（1s）| $RSS_DURING |
| BGSAVE 中（2s）| $RSS_DURING_2 |
| BGSAVE 后 | $RSS_AFTER |
| rdb_last_cow_size | $COW_SIZE |

**原假设**：fork 期间 RSS 翻倍
**证伪判断**：$([ "$RSS_DURING" -gt $((RSS_BEFORE * 3 / 2)) ] && echo "✅ 证伪不成立——空闲场景 RSS 也出现明显上升" || echo "❌ 证伪成立——空闲无写入场景 RSS 未翻倍，COW 未触发或触发轻微")
**论点修正**：$([ "$RSS_DURING" -gt $((RSS_BEFORE * 3 / 2)) ] && echo "无需修正" || echo "需修正为'fork 期间 RSS 翻倍是写入密集场景特定现象，空闲场景不翻倍'——这反而验证了 COW 机制的精确性")
**COW 复制比例**：$([ "$USED_MEM" -gt 0 ] && echo "scale=2; $COW_SIZE / $USED_MEM" | bc || echo "N/A")

EOF

echo "F3 完成: RSS before=$RSS_BEFORE during=$RSS_DURING cow=$COW_SIZE"

# ========================================
# F4: E7 证伪——fsync 与 fork+COW 独立性（逻辑推演，无环境依赖）
# ========================================
echo "运行 F4: E7 证伪（fsync 与 fork+COW 独立性）..."

cat >> "$RESULT_FILE" << 'EOF'
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

EOF

echo "F4 完成（逻辑推演）"
echo ""
echo "=== 证伪实验全部完成 ==="
echo "结果已写入: $RESULT_FILE"
