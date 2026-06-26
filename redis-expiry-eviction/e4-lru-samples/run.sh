#!/bin/bash
# E4 实验：maxmemory-samples 对 LRU 近似效果的影响（动态计算版）
# 环境：Redis 7.0-alpine (Docker), redis-cli 8.8.0
# 策略：先创建 1000 个 key，测量实际内存，设置 maxmemory 为 70% 来强制淘汰约 30%

echo "=========================================="
echo "E4 实验：maxmemory-samples 对 LRU 近似效果"
echo "环境：Redis 7.0 (Docker)"
echo "=========================================="

REDIS_CLI="redis-cli -p 16379"

# 先测量空数据库和满数据库的内存
$REDIS_CLI FLUSHDB
EMPTY=$($REDIS_CLI INFO memory | grep "used_memory:" | awk -F: '{print $2}' | tr -d '\r')
for i in $(seq 1 1000); do
    $REDIS_CLI SET "m:$i" "v" > /dev/null 2>&1
done
FULL=$($REDIS_CLI INFO memory | grep "used_memory:" | awk -F: '{print $2}' | tr -d '\r')
# 保留 60% 的内存容量 -> 淘汰约 40%
TARGET=$(( EMPTY + (FULL - EMPTY) * 60 / 100 ))
echo "空数据库：${EMPTY}B，1000 key：${FULL}B"
echo "目标 maxmemory：${TARGET}B（60% 容量）"

run_test() {
    local LABEL=$1
    local SAMPLES=$2
    
    echo ""
    echo "=== 实验：samples=$SAMPLES ==="
    
    $REDIS_CLI FLUSHDB
    $REDIS_CLI CONFIG SET maxmemory-policy allkeys-lru
    $REDIS_CLI CONFIG SET maxmemory-samples $SAMPLES
    $REDIS_CLI CONFIG SET maxmemory 0
    
    for i in $(seq 1 1000); do
        $REDIS_CLI SET "${LABEL}:$i" "v" > /dev/null 2>&1
    done
    
    # 访问前 200 个 key，每个访问 10 次
    for i in $(seq 1 200); do
        for j in $(seq 1 10); do
            $REDIS_CLI GET "${LABEL}:$i" > /dev/null 2>&1
        done
    done
    
    $REDIS_CLI CONFIG SET maxmemory $TARGET
    sleep 2
    
    HOT=0
    for i in $(seq 1 200); do
        R=$($REDIS_CLI EXISTS "${LABEL}:$i")
        HOT=$((HOT + R))
    done
    
    COLD=0
    for i in $(seq 801 1000); do
        R=$($REDIS_CLI EXISTS "${LABEL}:$i")
        COLD=$((COLD + R))
    done
    
    SIZE=$($REDIS_CLI DBSIZE | tr -d '\r')
    echo "samples=$SAMPLES：热 key (1-200) 剩余 $HOT/200，冷 key (801-1000) 剩余 $COLD/200，DBSIZE=$SIZE"
}

run_test "def" 5
run_test "high" 20
run_test "low" 1

echo ""
echo "=== 恢复默认配置 ==="
$REDIS_CLI CONFIG SET maxmemory-samples 5
$REDIS_CLI CONFIG SET maxmemory 0

echo ""
echo "=========================================="
echo "E4 实验完成"
echo "=========================================="
