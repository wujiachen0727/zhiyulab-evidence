#!/bin/bash
# E3 实验 Part 2：独立跑 Group B 时间序列采样
#
# 原因：Part 1（run-experiment.sh）的 Group B 全部失败——
# 因为 Go CPU profile 是全局互斥的，Group A 的 30s 采样占用期间，
# Group B 的 6 次短采样全部返回 "cpu profiling already in use"。
#
# 这本身是一个发现：**pprof 的架构决定了"传统 pprof"和"持续 profiling"
# 天生不能并发跑**。持续 profiling 工具用的是完全不同的架构（外部采样、
# 时间切片聚合）。
#
# Part 2 的做法：只跑 Group B（每 5 秒一次 5 秒采样，共 6 次），
# 压测流量和 Part 1 相同，毛刺窗口仍然是 15-20s（服务端决定）。

set -e

cd "$(dirname "$0")"
mkdir -p output

echo "=== E3 Part 2：Group B 时间序列采样 ==="
echo "实验时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 清掉前一次失败的 Group B 文件
rm -f output/groupB-window-*.pprof

echo "[1/3] 启动 HTTP 服务..."
go run main.go > output/server-part2.log 2>&1 &
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

echo "[3/3] 启动压测 + 时间序列采样（共 30s）..."
echo ""

# 压测配置
# - /fast：高 QPS，模拟正常流量背景
# - /slow：低 QPS，毛刺窗口内每秒约 10 次慢请求（50 次 × 200ms ≈ 10s CPU 压力）
hey -n 30000 -c 50 -z 30s -q 1000 http://localhost:6060/fast > output/hey-fast-part2.log 2>&1 &
HEY_FAST_PID=$!

# /slow：-q 10 限速每个并发每秒 10 次；-c 1 并发=1；总限 10 QPS
hey -c 1 -z 30s -q 10 http://localhost:6060/slow > output/hey-slow-part2.log 2>&1 &
HEY_SLOW_PID=$!

# 6 次连续 5 秒窗口，覆盖 0-30s
for i in 1 2 3 4 5 6; do
  START_AT=$((($i-1)*5))
  END_AT=$(($i*5))
  echo "    [窗口 $i] 覆盖 ${START_AT}-${END_AT}s 窗口，采样 5s..."
  curl -s -o "output/groupB-window-${i}.pprof" \
    "http://localhost:6060/debug/pprof/profile?seconds=5"
done

wait $HEY_FAST_PID 2>/dev/null || true
wait $HEY_SLOW_PID 2>/dev/null || true

echo ""
echo "=== 服务最终状态 ==="
curl -s http://localhost:6060/stats | tee output/final-stats-part2.txt

echo ""
echo "=== Group B 文件 ==="
ls -la output/groupB-window-*.pprof | awk '{print $NF, "("$5" bytes)"}'

echo ""
echo "=== Part 2 完成 ==="