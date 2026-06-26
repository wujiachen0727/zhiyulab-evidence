#!/bin/bash
# E2 实验：定期删除（Active Expiry Cycle）行为验证
# 环境：Redis 7.0-alpine (Docker), redis-cli 8.8.0
# 目的：验证 Redis 后台定期删除机制——批量设置大量相同 TTL 的 key，
#       观察 expired_keys 指标的变化

echo "=========================================="
echo "E2 实验：定期删除（Active Expiry Cycle）"
echo "环境：Redis 7.0 (Docker)"
echo "=========================================="

REDIS_CLI="redis-cli -p 16379"

echo ""
echo "=== 步骤 1：记录初始状态 ==="
$REDIS_CLI INFO stats | grep "expired_keys"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 2：批量写入 1000 个带 TTL 的 key ==="
# 用 pipeline 加速写入
for i in $(seq 1 1000); do
    echo "SET key:$i \"value:$i\""
    echo "EXPIRE key:$i 5"
done | $REDIS_CLI --pipe > /dev/null 2>&1
echo "已写入 1000 个 key（TTL=5秒）"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 3：观察定期删除（等待 8 秒让后台线程跑）==="
# 主动用 GET 访问其中 10 个 key（模拟惰性删除触发），其余靠定期删除
for i in $(seq 1 10); do
    $REDIS_CLI GET "key:$i" > /dev/null 2>&1
done
echo "已主动访问前 10 个 key"

sleep 8
echo "等待 8 秒后..."

echo ""
echo "=== 步骤 4：查看定期删除效果 ==="
$REDIS_CLI INFO stats | grep "expired_keys"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 5：检查剩余 key 是否都被清除了 ==="
REMAINING=$($REDIS_CLI DBSIZE)
echo "当前数据库 key 数量：$REMAINING"
echo "定期删除清理了 $((1000 - REMAINING + 10)) 个 key（含惰性删除的 10 个）"

echo ""
echo "=== 步骤 6：大容量测试——写入 10000 个 key 观察定期删除压力 ==="
$REDIS_CLI FLUSHDB
INIT_EXPIRED=$($REDIS_CLI INFO stats | grep "expired_keys" | awk -F: '{print $2}' | tr -d '\r')

for i in $(seq 1 10000); do
    echo "SET bulk:$i \"data:$i\""
    echo "EXPIRE bulk:$i 5"
done | $REDIS_CLI --pipe > /dev/null 2>&1
echo "已写入 10000 个 key（TTL=5秒）"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 7：等待 10 秒观察定期删除处理大量过期 key 的效果 ==="
sleep 10
AFTER_EXPIRED=$($REDIS_CLI INFO stats | grep "expired_keys" | awk -F: '{print $2}' | tr -d '\r')
echo "定期删除处理了 $((AFTER_EXPIRED - INIT_EXPIRED)) 个过期 key"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 8：极端测试——全部 key 同时过期 ==="
$REDIS_CLI FLUSHDB
INIT_EXPIRED2=$($REDIS_CLI INFO stats | grep "expired_keys" | awk -F: '{print $2}' | tr -d '\r')

for i in $(seq 1 50000); do
    echo "SET flood:$i \"payload:$i\""
    echo "EXPIRE flood:$i 5"
done | $REDIS_CLI --pipe > /dev/null 2>&1
echo "已写入 50000 个 key（TTL=5秒）"
$REDIS_CLI DBSIZE

echo ""
echo "=== 步骤 9：等待 15 秒 ==="
sleep 15
AFTER_EXPIRED2=$($REDIS_CLI INFO stats | grep "expired_keys" | awk -F: '{print $2}' | tr -d '\r')
echo "定期删除处理了 $((AFTER_EXPIRED2 - INIT_EXPIRED2)) 个过期 key"
$REDIS_CLI DBSIZE

echo ""
echo "=========================================="
echo "E2 实验完成"
echo "=========================================="
