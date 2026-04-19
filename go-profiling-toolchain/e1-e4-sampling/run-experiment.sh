#!/bin/bash
# E1 + E4 实验脚本：三种采样率跑同一程序
#
# 预期：
#   - 100Hz（Go 默认）：heavy 和 medium 清晰可见；short 被低估；micro 几乎消失
#   - 1000Hz：short 明显变清晰；micro 开始可见
#   - 10000Hz：所有函数都能看见（但 Go 运行时可能限制实际采样率）

set -e

cd "$(dirname "$0")"
mkdir -p output

echo "=== E1 + E4 采样基线 + 采样频率对比 ==="
echo "实验时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "环境: Go $(go version | awk '{print $3}') / $(uname -srm)"
echo ""

# 编译
echo "[0/3] 编译..."
go build -o workload main.go
echo "    ✅ 编译完成"
echo ""

# 三种采样率
for hz in 100 1000 10000; do
  echo "=== 运行 ${hz}Hz ==="
  RATE_HZ=$hz ./workload 2>&1 | tee output/run-${hz}hz.log
  echo ""
done

# 分析
echo "=== Profile 分析 ==="
for hz in 100 1000 10000; do
  echo ""
  echo "----- ${hz}Hz -----"
  go tool pprof -top -nodecount=10 output/cpu-${hz}hz.pprof 2>&1 | head -20
done | tee output/e1-e4-comparison.txt

echo ""
echo "=== 实验完成 ==="
echo "关键文件："
ls -la output/*.pprof output/*.log output/*.txt 2>/dev/null | awk '{print "  " $NF}'