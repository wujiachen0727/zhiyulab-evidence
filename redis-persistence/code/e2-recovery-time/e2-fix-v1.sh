#!/bin/bash
# E2 修复：用轮询方式采集 loading_total_time
set -e

cd "$(dirname "$0")/../../.."

# 准备 RDB 文件（100万key）
docker rm -f e2-fix-prep 2>/dev/null || true
docker run -d --name e2-fix-prep -p 6420:6379 \
  redis:7-alpine redis-server --save "" --appendonly no --dbfilename dump.rdb --dir /data
sleep 2
docker exec e2-fix-prep sh -c 'for i in $(seq 1 1000000); do echo "SET key_$i value_$i"; done | redis-cli --pipe'
sleep 2
docker exec e2-fix-prep redis-cli BGSAVE
sleep 3
while [ "$(docker exec e2-fix-prep redis-cli INFO persistence | grep rdb_bgsave_in_progress | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')" = "1" ]; do sleep 1; done
docker cp e2-fix-prep:/data/dump.rdb /tmp/e2-fix-dump.rdb
docker rm -f e2-fix-prep

# 启动 Redis 加载 RDB
docker rm -f e2-fix-load 2>/dev/null || true
docker run -d --name e2-fix-load -p 6421:6379 \
  -v /tmp/e2-fix-dump.rdb:/data/dump.rdb \
  redis:7-alpine redis-server --save "" --appendonly no --dbfilename dump.rdb --dir /data

# 轮询 loading 状态
RDB_LOADING_TIME=""
RDB_LOADED_KEYS=""
for i in $(seq 1 20); do
  LOADING=$(docker exec e2-fix-load redis-cli INFO persistence 2>/dev/null | grep '^loading:' | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ' || echo "")
  if [ "$LOADING" = "0" ]; then
    RDB_LOADING_TIME=$(docker exec e2-fix-load redis-cli INFO persistence 2>/dev/null | grep loading_total_time | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')
    RDB_LOADED_KEYS=$(docker exec e2-fix-load redis-cli DBSIZE 2>/dev/null | tr -d '\r')
    break
  fi
  sleep 0.3
done
echo "RDB: loading_total_time=${RDB_LOADING_TIME}μs, keys=${RDB_LOADED_KEYS}"
docker rm -f e2-fix-load

# 准备 AOF 文件
docker rm -f e2-fix-aof-prep 2>/dev/null || true
docker run -d --name e2-fix-aof-prep -p 6422:6379 \
  redis:7-alpine redis-server --appendonly yes --appendfsync everysec --save "" --dir /data
sleep 2
docker exec e2-fix-aof-prep sh -c 'for i in $(seq 1 1000000); do echo "SET key_$i value_$i"; done | redis-cli --pipe'
sleep 2
docker exec e2-fix-aof-prep redis-cli BGREWRITEAOF
sleep 3
while [ "$(docker exec e2-fix-aof-prep redis-cli INFO persistence | grep aof_rewrite_in_progress | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')" = "1" ]; do sleep 1; done
docker cp e2-fix-aof-prep:/data/appendonlydir /tmp/e2-fix-appendonlydir
docker rm -f e2-fix-aof-prep

# 启动 Redis 加载 AOF
docker rm -f e2-fix-aof-load 2>/dev/null || true
docker run -d --name e2-fix-aof-load -p 6423:6379 \
  -v /tmp/e2-fix-appendonlydir:/data/appendonlydir \
  redis:7-alpine redis-server --appendonly yes --appendfsync everysec --save "" --dir /data

AOF_LOADING_TIME=""
AOF_LOADED_KEYS=""
for i in $(seq 1 20); do
  LOADING=$(docker exec e2-fix-aof-load redis-cli INFO persistence 2>/dev/null | grep '^loading:' | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ' || echo "")
  if [ "$LOADING" = "0" ]; then
    AOF_LOADING_TIME=$(docker exec e2-fix-aof-load redis-cli INFO persistence 2>/dev/null | grep loading_total_time | tr -d '\r' | awk -F: '{print $2}' | tr -d ' ')
    AOF_LOADED_KEYS=$(docker exec e2-fix-aof-load redis-cli DBSIZE 2>/dev/null | tr -d '\r')
    break
  fi
  sleep 0.3
done
echo "AOF: loading_total_time=${AOF_LOADING_TIME}μs, keys=${AOF_LOADED_KEYS}"
docker rm -f e2-fix-aof-load

# 输出汇总
echo ""
echo "=== E2 修复结果 ==="
echo "RDB: ${RDB_LOADING_TIME}μs (${RDB_LOADED_KEYS} keys)"
echo "AOF: ${AOF_LOADING_TIME}μs (${AOF_LOADED_KEYS} keys)"
if [ -n "$RDB_LOADING_TIME" ] && [ -n "$AOF_LOADING_TIME" ] && [ "$AOF_LOADING_TIME" -gt 0 ]; then
  echo "RDB/AOF 比值: $(echo "scale=2; $RDB_LOADING_TIME / $AOF_LOADING_TIME" | bc)"
fi
