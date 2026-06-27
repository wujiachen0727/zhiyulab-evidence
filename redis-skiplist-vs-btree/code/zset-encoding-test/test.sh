#!/bin/bash
# E1: 实测 ZSET 编码切换（listpack -> skiplist）
# Redis 8.8.0, 默认配置

REDIS_PORT=16379
redis-server --port $REDIS_PORT --daemonize yes --save "" --appendonly no 
sleep 0.5

echo "=== Redis 8.8.0 ZSET 编码切换实测 ==="
echo "时间: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo ""

# 测试 1: 逐渐增加元素数量，观察编码切换
echo "=== 测试 1: 元素数量阈值 ==="
echo "| 元素数 | 元素大小 | encoding | memory(bytes) |"
echo "|-------:|---------:|----------|--------------:|"

for count in 1 64 127 128 129 200; do
  redis-cli -p $REDIS_PORT DEL testkey >/dev/null
  # 用短成员（< 64 字节）
  for i in $(seq 1 $count); do
    redis-cli -p $REDIS_PORT ZADD testkey $i "member$i" >/dev/null 2>&1
  done
  enc=$(redis-cli -p $REDIS_PORT OBJECT ENCODING testkey)
  mem=$(redis-cli -p $REDIS_PORT MEMORY USAGE testkey)
  echo "| $count | ~10B | $enc | $mem |"
done

echo ""
echo "=== 测试 2: 单个元素大小阈值 ==="
echo "| 元素数 | 元素大小 | encoding | memory(bytes) |"
echo "|-------:|---------:|----------|--------------:|"

# 测试单个元素大小超 64 字节触发切换
redis-cli -p $REDIS_PORT DEL testkey2 >/dev/null
# 64 字节成员
big_member=$(python3 -c "print('a'*64)")
redis-cli -p $REDIS_PORT ZADD testkey2 1 "$big_member" >/dev/null
enc=$(redis-cli -p $REDIS_PORT OBJECT ENCODING testkey2)
mem=$(redis-cli -p $REDIS_PORT MEMORY USAGE testkey2)
echo "| 1 | 64B | $enc | $mem |"

redis-cli -p $REDIS_PORT DEL testkey2 >/dev/null
# 65 字节成员
big_member=$(python3 -c "print('a'*65)")
redis-cli -p $REDIS_PORT ZADD testkey2 1 "$big_member" >/dev/null
enc=$(redis-cli -p $REDIS_PORT OBJECT ENCODING testkey2)
mem=$(redis-cli -p $REDIS_PORT MEMORY USAGE testkey2)
echo "| 1 | 65B | $enc | $mem |"

echo ""
echo "=== 测试 3: listpack vs skiplist 内存占用对比 ==="
echo "| 元素数 | encoding | memory(bytes) | per-element(bytes) |"
echo "|-------:|----------|--------------:|-------------------:|"

for count in 100 128 129 500 1000; do
  redis-cli -p $REDIS_PORT DEL testkey3 >/dev/null
  for i in $(seq 1 $count); do
    redis-cli -p $REDIS_PORT ZADD testkey3 $i "member$i" >/dev/null 2>&1
  done
  enc=$(redis-cli -p $REDIS_PORT OBJECT ENCODING testkey3)
  mem=$(redis-cli -p $REDIS_PORT MEMORY USAGE testkey3)
  per=$(python3 -c "print(round($mem/$count, 1))")
  echo "| $count | $enc | $mem | $per |"
done

echo ""
echo "=== 配置确认 ==="
echo "zset-max-listpack-entries: $(redis-cli -p $REDIS_PORT CONFIG GET zset-max-listpack-entries | tail -1)"
echo "zset-max-listpack-value: $(redis-cli -p $REDIS_PORT CONFIG GET zset-max-listpack-value | tail -1)"

redis-cli -p $REDIS_PORT SHUTDOWN NOSAVE >/dev/null 2>&1
