#!/bin/bash
# E3 偶发毛刺证伪实验脚本
#
# 实验流程：
#   1. 启动 HTTP 服务（背景运行）
#   2. 等服务就绪
#   3. 启动压测（99% /fast + 1% /slow），共 30 秒
#   4. 同时采集两组 profile：
#      - Group A：一次性 30 秒 CPU profile（传统 pprof 方式）
#      - Group B：每 5 秒一次短窗口 profile，共 6 次（模拟持续 profiling）
#   5. 输出采样结果到 output/
#
# 预期：
#   - Group A：慢路径占比被稀释
#   - Group B：第 3-4 个窗口（时间 15-20s）能清晰看到慢路径占比飙高

set -e

cd "$(dirname "$0")"
mkdir -p output

echo "=== E3 偶发毛刺证伪实验 ==="
echo "实验时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "环境: Go $(go version | awk '{print $3}') / $(uname -srm)"
echo ""

# Step 1: 启动服务（后台）
echo "[1/4] 启动 HTTP 服务..."
go run main.go > output/server.log 2>&1 &
SERVER_PID=$!
echo "    Server PID: $SERVER_PID"

# 确保退出时清理
trap "echo '清理: 停止服务 $SERVER_PID'; kill $SERVER_PID 2>/dev/null; exit" INT TERM EXIT

# Step 2: 等服务就绪
echo "[2/4] 等待服务就绪..."
for i in {1..10}; do
  if curl -s http://localhost:6060/stats >/dev/null 2>&1; then
    echo "    服务就绪"
    break
  fi
  sleep 0.5
done

# Step 3: 启动压测 + 并行采样
echo "[3/4] 启动压测 + 采样..."
echo ""

# 3a: 后台启动 Group A——单次 30 秒 profile
echo "    [Group A] 一次性 30 秒 CPU profile（传统 pprof 方式）..."
curl -s -o output/groupA-30s.pprof \
  "http://localhost:6060/debug/pprof/profile?seconds=30" &
GROUPA_PID=$!

# 3b: 压测（99% /fast + 1% /slow）
# 策略：起两个 hey，比例 99:1
# /fast：总 30000 请求、30 秒内发完、并发 50
# /slow：总 300 请求、30 秒内发完、并发 5
#
# 注意：slow 的"真正慢"只在 15-20s 窗口内触发（服务端决定）
echo "    [压测] 启动 /fast（99%）+ /slow（1%）并行压测..."
hey -n 30000 -c 50 -z 30s -q 1000 http://localhost:6060/fast > output/hey-fast.log 2>&1 &
HEY_FAST_PID=$!

# /slow：-q 10 × -c 1 并发 = 10 QPS，毛刺窗口 5s × 10 × 200ms = 10s CPU 压力
hey -c 1 -z 30s -q 10 http://localhost:6060/slow > output/hey-slow.log 2>&1 &
HEY_SLOW_PID=$!

# 3c: Group B——每 5 秒一次短窗口 profile（共 6 次）
echo "    [Group B] 每 5 秒一次 5 秒 profile（模拟持续 profiling 的时间序列）..."
for i in 1 2 3 4 5 6; do
  echo "      [窗口 $i] 采样中（约 5s）..."
  # 每次阻塞 5s 采样，采样结束后立即开始下一次
  curl -s -o "output/groupB-window-${i}.pprof" \
    "http://localhost:6060/debug/pprof/profile?seconds=5"
done

echo ""
echo "[4/4] 等待所有任务完成..."

# 确保 Group A 的 30 秒采样已完成
wait $GROUPA_PID 2>/dev/null || true
wait $HEY_FAST_PID 2>/dev/null || true
wait $HEY_SLOW_PID 2>/dev/null || true

# 抓取最终服务状态
echo ""
echo "=== 服务最终状态 ==="
curl -s http://localhost:6060/stats | tee output/final-stats.txt

echo ""
echo "=== 生成的文件 ==="
ls -la output/*.pprof 2>/dev/null | awk '{print $NF, "("$5" bytes)"}'

echo ""
echo "=== 实验完成 ==="
echo "下一步：分析两组 profile，对比慢路径占比"