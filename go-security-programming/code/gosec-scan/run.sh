#!/usr/bin/env bash
# E1: gosec 扫描 10 个主流 Go 开源项目
#
# 用途：统计真实 Go 项目中 gosec 高危发现的分布，验证"第二层占大头"论断。
#
# 前提：
#   - Go 1.22+
#   - gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest
#   - 约 2GB 磁盘空间（克隆 10 个仓库）
#
# 运行时间：约 2-5 分钟（首次克隆 + 扫描）

set -e

WORKDIR="${WORKDIR:-/tmp/gosec-scan-targets}"
OUTDIR="${OUTDIR:-/tmp/gosec-results}"

mkdir -p "$WORKDIR" "$OUTDIR"
cd "$WORKDIR"

# 被扫描的 10 个项目
REPOS=(
  "gin-gonic/gin"
  "labstack/echo"
  "gorilla/mux"
  "spf13/cobra"
  "spf13/viper"
  "stretchr/testify"
  "sirupsen/logrus"
  "go-redis/redis"
  "grpc/grpc-go"
  "prometheus/client_golang"
)

echo ">>> 克隆/更新项目"
for repo in "${REPOS[@]}"; do
  name=$(basename "$repo")
  if [ -d "$name" ]; then
    echo "  $name 已存在"
  else
    echo "  克隆 $repo"
    git clone --depth 1 "https://github.com/$repo.git" "$name" 2>&1 | tail -1
  fi
done

echo
echo ">>> 扫描"
for repo in "${REPOS[@]}"; do
  name=$(basename "$repo")
  echo "  扫描 $name"
  (cd "$name" && gosec -quiet -fmt=json -out="$OUTDIR/gosec-$name.json" ./... 2>&1 | tail -3)
done

echo
echo ">>> 完成。原始 JSON 存放于 $OUTDIR/"
echo "   运行 aggregate.py 聚合数据。"
