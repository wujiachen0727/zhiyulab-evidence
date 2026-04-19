#!/bin/bash
# E5 Pyroscope 开销实测脚本
#
# 实验流程：
#   1. 确保 Pyroscope server 正在运行（localhost:4040）
#   2. 启动基线服务（无 agent，端口 6061），压测 60 秒
#   3. 启动 agent 服务（挂 Pyroscope Go SDK，端口 6062），压测 60 秒
#   4. 对比两组 QPS、延迟、进程 CPU/RSS
#
# 预期：
#   - 基线 QPS 明显高于 agent 版（Grafana 声称 <1%）
#   - 进程 RSS 略有提升（Pyroscope agent 维护 profile 缓冲区）

set -e

cd "$(dirname "$0")"
mkdir -p output

echo "=== E5 Pyroscope 开销实测 ==="
echo "实验时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "环境: Go $(go version | awk '{print $3}') / $(uname -srm)"
echo ""

# Step 0: 确认 Pyroscope server 就绪
echo "[0/3] 验证 Pyroscope server..."
if ! curl -s -f http://localhost:4040/ready >/dev/null 2>&1; then
  echo "❌ Pyroscope server 未就绪 (http://localhost:4040/ready)"
  echo "请先运行: docker run -d --name pyroscope-e5 -p 4040:4040 grafana/pyroscope:latest"
  exit 1
fi
echo "    ✅ Pyroscope server 就绪"

# 采集函数：给定 PID，周期性记录 CPU/RSS 到文件
collect_proc_stats() {
  local pid=$1
  local outfile=$2
  local duration=$3  # 秒
  local interval=5

  echo "timestamp,cpu_pct,rss_kb" > "$outfile"
  local start_ts=$(date +%s)
  while [ $(($(date +%s) - start_ts)) -lt $duration ]; do
    if ! kill -0 $pid 2>/dev/null; then
      echo "    ⚠️ 进程 $pid 已退出"
      break
    fi
    # ps 输出格式：CPU% 和 RSS (KB)
    ps -p $pid -o %cpu,rss 2>/dev/null | awk -v ts=$(date +%s) 'NR>1 {printf "%s,%s,%s\n", ts, $1, $2}' >> "$outfile"
    sleep $interval
  done
}

# ============================================================
# Group A: 基线（无 agent）
# ============================================================
echo ""
echo "[1/3] Group A: 基线服务（无 Pyroscope agent）"
cd baseline
# 用 go build 编译后直接运行，这样 PID 就是服务本身（而不是 go run 的外层）
go build -o ../output/baseline-bin main.go
../output/baseline-bin > ../output/baseline-server.log 2>&1 &
BASELINE_PID=$!
cd ..
echo "    PID: $BASELINE_PID"

trap "echo '清理: 停止所有 Go 进程'; kill $BASELINE_PID 2>/dev/null; [ -n \"$AGENT_PID\" ] && kill $AGENT_PID 2>/dev/null; exit" INT TERM EXIT

# 等服务就绪
sleep 2
for i in {1..10}; do
  if curl -s http://localhost:6061/stats >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

echo "    启动进程监控..."
collect_proc_stats $BASELINE_PID output/baseline-proc.csv 65 &
MON_A_PID=$!

echo "    压测 60 秒（-c 30 -z 60s，不限 QPS）..."
hey -c 30 -z 60s http://localhost:6061/work > output/baseline-hey.log 2>&1

# 等监控完成
wait $MON_A_PID 2>/dev/null || true

# 最终状态
echo "    服务最终状态:"
curl -s http://localhost:6061/stats | tee output/baseline-final-stats.txt
echo ""

kill $BASELINE_PID 2>/dev/null || true
wait $BASELINE_PID 2>/dev/null || true

sleep 2  # 等端口释放

# ============================================================
# Group B: 带 agent（Pyroscope push mode）
# ============================================================
echo ""
echo "[2/3] Group B: agent 服务（挂 Pyroscope Go SDK）"
cd with-agent
go build -o ../output/agent-bin main.go
../output/agent-bin > ../output/agent-server.log 2>&1 &
AGENT_PID=$!
cd ..
echo "    PID: $AGENT_PID"

# 等服务就绪
sleep 3
for i in {1..15}; do
  if curl -s http://localhost:6062/stats >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

echo "    启动进程监控..."
collect_proc_stats $AGENT_PID output/agent-proc.csv 65 &
MON_B_PID=$!

echo "    压测 60 秒（-c 30 -z 60s，不限 QPS）..."
hey -c 30 -z 60s http://localhost:6062/work > output/agent-hey.log 2>&1

wait $MON_B_PID 2>/dev/null || true

echo "    服务最终状态:"
curl -s http://localhost:6062/stats | tee output/agent-final-stats.txt
echo ""

kill $AGENT_PID 2>/dev/null || true
wait $AGENT_PID 2>/dev/null || true

# ============================================================
# Step 3: 提取核心指标
# ============================================================
echo ""
echo "[3/3] 对比提取..."

# 提取 hey 输出里的关键指标
extract_hey_metrics() {
  local file=$1
  local label=$2
  echo "=== $label ==="
  grep -E "Total:|Slowest:|Fastest:|Average:|Requests/sec:" "$file"
  echo ""
  echo "Latency distribution:"
  sed -n '/Latency distribution:/,/Details/p' "$file" | head -10
  echo ""
}

{
  echo "================================"
  echo "E5 Pyroscope 开销实测最终对比"
  echo "================================"
  echo ""
  extract_hey_metrics output/baseline-hey.log "基线（无 agent）"
  extract_hey_metrics output/agent-hey.log "带 agent（Pyroscope push mode）"

  echo "=== 进程资源（平均）==="
  echo ""
  echo "基线："
  awk -F, 'NR>1 {cpu+=$2; rss+=$3; n++} END {printf "  平均 CPU: %.1f%%\n  平均 RSS: %.1f MB\n  样本数: %d\n", cpu/n, rss/n/1024, n}' output/baseline-proc.csv
  echo ""
  echo "带 agent："
  awk -F, 'NR>1 {cpu+=$2; rss+=$3; n++} END {printf "  平均 CPU: %.1f%%\n  平均 RSS: %.1f MB\n  样本数: %d\n", cpu/n, rss/n/1024, n}' output/agent-proc.csv
} | tee output/e5-comparison.txt

echo ""
echo "=== 实验完成 ==="
echo "关键文件："
ls -la output/*.log output/*.csv output/*.txt 2>/dev/null | awk '{print "  " $NF}'