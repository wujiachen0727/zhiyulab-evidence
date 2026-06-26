#!/bin/bash
# E1 实验：惰性删除（Lazy Expiry）触发时机验证
# 环境：Redis 7.0-alpine (Docker), redis-cli 8.8.0
# 目的：验证设置了 TTL 的 key 在过期后不会立即删除，只有在被访问时才触发惰性删除

echo "=========================================="
echo "E1 实验：惰性删除触发时机"
echo "环境：Redis 7.0 (Docker)"
echo "=========================================="

REDIS_CLI="redis-cli -p 16379"

echo ""
echo "=== 步骤 1：清空当前数据库 ==="
$REDIS_CLI FLUSHDB

echo ""
echo "=== 步骤 2：设置带 TTL 的 key ==="
$REDIS_CLI SET mykey "hello"
$REDIS_CLI EXPIRE mykey 5
echo "设置了 mykey（TTL=5秒）"

echo ""
echo "=== 步骤 3：立即查看 TTL ==="
$REDIS_CLI TTL mykey

echo ""
echo "=== 步骤 4：等待 6 秒（超过过期时间）==="
sleep 6
echo "等待 6 秒后..."

echo ""
echo "=== 步骤 5：直接 GET 访问（触发惰性删除）==="
$REDIS_CLI GET mykey

echo ""
echo "=== 步骤 6：验证 key 是否已被删除 ==="
$REDIS_CLI EXISTS mykey

echo ""
echo "=== 步骤 7：设置新 key 并查看过期后的内存占用 ==="
$REDIS_CLI SET lazykey "i-will-expire"
$REDIS_CLI EXPIRE lazykey 3
echo "设置了 lazykey（TTL=3秒）"

echo ""
echo "=== 步骤 8：在过期前查看 key 是否存在 ==="
$REDIS_CLI EXISTS lazykey

echo ""
echo "=== 步骤 9：等待 4 秒（超过过期时间）==="
sleep 4
echo "等待 4 秒后..."

echo ""
echo "=== 步骤 10：不访问 key，直接检查 EXISTS ==="
# EXISTS 不会触发惰性删除，惰性删除只在读取/写入操作时触发
$REDIS_CLI EXISTS lazykey

echo ""
echo "=== 步骤 11：用 GET 访问后再检查 ==="
$REDIS_CLI GET lazykey
$REDIS_CLI EXISTS lazykey

echo ""
echo "=== 步骤 12：演示惰性删除的实际效果——过期 key 占用内存 ==="
$REDIS_CLI SET "bigdata" "$(python3 -c "print('x'*10000)")"
$REDIS_CLI EXPIRE bigdata 2
echo "设置了 bigdata（10KB 数据，TTL=2秒）"

echo ""
echo "=== 步骤 13：使用 INFO 查看内存（过期前）==="
$REDIS_CLI INFO memory | grep -E "used_memory_human|used_memory"

echo ""
echo "=== 步骤 14：等待 3 秒 ==="
sleep 3
echo "等待 3 秒后（key 已过期但未被访问）..."

echo ""
echo "=== 步骤 15：过期后但未访问时的内存 ==="
$REDIS_CLI INFO memory | grep -E "used_memory_human|used_memory"

echo ""
echo "=== 步骤 16：访问该 key（触发惰性删除）==="
$REDIS_CLI GET bigdata

echo ""
echo "=== 步骤 17：访问后再次查看内存 ==="
$REDIS_CLI INFO memory | grep -E "used_memory_human|used_memory"

echo ""
echo "=========================================="
echo "E1 实验完成"
echo "=========================================="
