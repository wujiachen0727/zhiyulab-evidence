#!/bin/bash
# E3 实验：不同内存淘汰策略的行为对比
# 环境：Redis 7.0-alpine (Docker), redis-cli 8.8.0
# 目的：设置不同的 maxmemory-policy，观察淘汰行为差异

echo "=========================================="
echo "E3 实验：不同淘汰策略对比"
echo "环境：Redis 7.0 (Docker)"
echo "=========================================="

REDIS_CLI="redis-cli -p 16379"

# 重置：大内存避免干扰
$REDIS_CLI CONFIG SET maxmemory 0
$REDIS_CLI FLUSHDB

echo ""
echo "=== 实验 1：noeviction（不淘汰，写入报错） ==="
$REDIS_CLI CONFIG SET maxmemory 2097152  # 2MB
$REDIS_CLI CONFIG SET maxmemory-policy noeviction
echo "maxmemory=2MB, policy=noeviction"

# 写入数据，直到报错
for i in $(seq 1 200); do
    RESULT=$($REDIS_CLI SET "noevict:$i" "$(python3 -c "print('x'*10000)")" 2>&1)
    if echo "$RESULT" | grep -q "OOM"; then
        echo "noeviction 模式下第 $i 个 key 写入失败（OOM）——验证：noeviction 真的不淘汰"
        break
    fi
done

echo ""
echo "=== 实验 2：allkeys-lru（淘汰最近最少使用的 key） ==="
$REDIS_CLI CONFIG SET maxmemory-policy allkeys-lru
$REDIS_CLI FLUSHDB

# 先创建 50 个 key，访问其中 20 个让它们变"热"
for i in $(seq 1 50); do
    $REDIS_CLI SET "lru:$i" "$(python3 -c "print('y'*2000)")"
done
for i in $(seq 1 20); do
    $REDIS_CLI GET "lru:$i" > /dev/null
done
echo "已创建 50 个 key（每个 ~2KB），前 20 个是热 key"

# 设置小内存触发淘汰
$REDIS_CLI CONFIG SET maxmemory 131072  # 128KB
sleep 1

# 检查哪些 key 还在
echo "allkeys-lru 淘汰后..."
echo "热 key lru:1 是否存在：$($REDIS_CLI EXISTS lru:1)"
echo "热 key lru:10 是否存在：$($REDIS_CLI EXISTS lru:10)"
echo "热 key lru:20 是否存在：$($REDIS_CLI EXISTS lru:20)"
echo "冷 key lru:30 是否存在：$($REDIS_CLI EXISTS lru:30)"
echo "冷 key lru:40 是否存在：$($REDIS_CLI EXISTS lru:40)"
echo "冷 key lru:50 是否存在：$($REDIS_CLI EXISTS lru:50)"
echo "当前 DBSIZE：$($REDIS_CLI DBSIZE)"

echo ""
echo "=== 实验 3：volatile-ttl（淘汰剩余 TTL 最短的 key） ==="
$REDIS_CLI CONFIG SET maxmemory 0
$REDIS_CLI FLUSHDB

$REDIS_CLI CONFIG SET maxmemory-policy volatile-ttl
$REDIS_CLI SET "ttl_long" "data_long"
$REDIS_CLI EXPIRE "ttl_long" 3600
$REDIS_CLI SET "ttl_short" "data_short"
$REDIS_CLI EXPIRE "ttl_short" 60
$REDIS_CLI SET "ttl_shortest" "data_shortest"
$REDIS_CLI EXPIRE "ttl_shortest" 5

echo "已创建 3 个 key（不同 TTL）"
echo "  ttl_long TTL=$($REDIS_CLI TTL ttl_long)"
echo "  ttl_short TTL=$($REDIS_CLI TTL ttl_short)"
echo "  ttl_shortest TTL=$($REDIS_CLI TTL ttl_shortest)"

# 设置小内存触发淘汰
$REDIS_CLI CONFIG SET maxmemory 65536  # 64KB
sleep 2

echo "volatile-ttl 淘汰后..."
echo "ttl_long 是否存在：$($REDIS_CLI EXISTS ttl_long)"
echo "ttl_short 是否存在：$($REDIS_CLI EXISTS ttl_short)"
echo "ttl_shortest 是否存在：$($REDIS_CLI EXISTS ttl_shortest)"

echo ""
echo "=== 实验 4：volatile-lru vs allkeys-lru 的区别 ==="
$REDIS_CLI CONFIG SET maxmemory 0
$REDIS_CLI FLUSHDB

$REDIS_CLI CONFIG SET maxmemory-policy volatile-lru

# 创建 10 个有 TTL 的 key + 10 个无 TTL 的 key
for i in $(seq 1 10); do
    $REDIS_CLI SET "withttl:$i" "$(python3 -c "print('z'*5000)")"
    $REDIS_CLI EXPIRE "withttl:$i" 3600
    $REDIS_CLI SET "nottl:$i" "$(python3 -c "print('z'*5000)")"
done
echo "已创建 10 个有 TTL + 10 个无 TTL 的 key（每个 ~5KB）"

# 触发淘汰
$REDIS_CLI CONFIG SET maxmemory 131072  # 128KB
sleep 2

echo "volatile-lru 淘汰后（只能淘汰有 TTL 的 key）..."
echo "有 TTL 的 key 数量：$($REDIS_CLI DBSIZE | awk '{print $1}')"
# 检查无 TTL 的 key 是否还在
WITHTTL_CNT=0
NOTTL_CNT=0
for i in $(seq 1 10); do
    WITHTTL_CNT=$((WITHTTL_CNT + $($REDIS_CLI EXISTS withttl:$i)))
    NOTTL_CNT=$((NOTTL_CNT + $($REDIS_CLI EXISTS nottl:$i)))
done
echo "有 TTL 的 key 剩余：$WITHTTL_CNT"
echo "无 TTL 的 key 剩余：$NOTTL_CNT"

echo ""
echo "=========================================="
echo "E3 实验完成"
echo "=========================================="
