#!/bin/bash
# E3 实验 v2 Part 3：独立跑 Group A（30 秒单窗口）
#
# 为什么单独跑：
# v2 第一次运行中 Group A 和 Group B 的 curl 同时发起，哪个先到服务
# 就谁能获得 CPU profile 锁。本次实验里 Group B 赢了，Group A 被拒。
#
# 这本身就是文章可以用的一个点：你不能"同时"跑两种 pprof 方式——
# Go CPU profile 是全局单例的。
#
# 为了得到干净的 Group A 数据，这里独立跑一次。

set -e

cd "$(dirname "$0")"
mkdir -p output

echo "=== E3 Part 3：Group A 30 秒单窗口（独立运行）==="
echo "实验时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

echo "[1/3] 启动 HTTP 服务..."
go run main.go > output/server-part3.log 2>&1 &
SERVER_PID=$!
echo "    Server PID: $SERVER_PID"

trap "echo '清理: 停止服务 $SERVER_PID'; kill $SERVER_PID 2>/dev/null; exit" INT TERM EXIT

echo "[2/3] 等待服务就绪..."
for i in {1..10}; do
  if curl -s http://localhost:6060/stats >/dev/null 2>&1; then
    echo "    服务就绪"
    break
  fi
  sleep 0.5
done

echo "[3/3] 启动压测 + Group A 30 秒采样..."

# Group A：单次 30 秒采样（与 Part 2 的 Group B 流量配置一致）
curl -s -o output/groupA-30s.pprof "http://localhost:6060/debug/pprof/profile?seconds=30" &
GROUPA_PID=$!

# 压测配置与 Part 2 一致
hey -n 30000 -c 50 -z 30s -q 1000 http://localhost:6060/fast > output/hey-fast-part3.log 2>&1 &
HEY_FAST_PID=$!

hey -c 1 -z 30s -q 10 http://localhost:6060/slow > output/hey-slow-part3.log 2>&1 &
HEY_SLOW_PID=$!

# 等所有完成
wait $GROUPA_PID 2>/dev/null || true
wait $HEY_FAST_PID 2>/dev/null || true
wait $HEY_SLOW_PID 2>/dev/null || true

echo ""
echo "=== 服务最终状态 ==="
curl -s http://localhost:6060/stats | tee output/final-stats-part3.txt

echo ""
echo "=== Group A 文件大小 ==="
ls -la output/groupA-30s.pprof | awk '{print $NF, "("$5" bytes)"}'

echo ""
echo "=== Part 3 完成 ==="