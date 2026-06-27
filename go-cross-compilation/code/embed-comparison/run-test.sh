#!/bin/bash
# E3: embed 资源 vs 外置文件对比
# 测试不同大小资源嵌入后的 binary 体积变化

set -e
echo "=== E3: embed vs 外置文件对比 ==="
echo ""

TMPDIR=$(mktemp -d)
cd "$TMPDIR"

cat > go.mod << 'GOMOD'
module embed-test
go 1.22
GOMOD

# 生成不同大小的测试文件
echo "生成测试资源文件..."
dd if=/dev/urandom bs=1024 count=100 of=resource_100k.bin 2>/dev/null   # 100KB
dd if=/dev/urandom bs=1024 count=1024 of=resource_1m.bin 2>/dev/null    # 1MB
dd if=/dev/urandom bs=1024 count=5120 of=resource_5m.bin 2>/dev/null    # 5MB
dd if=/dev/urandom bs=1024 count=10240 of=resource_10m.bin 2>/dev/null  # 10MB

# 基准（无资源嵌入）
cat > main_base.go << 'GO'
package main
import "fmt"
func main() { fmt.Println("no embed") }
GO

CGO_ENABLED=0 go build -ldflags='-s -w' -o base main_base.go 2>/dev/null
base_size=$(ls -l base | awk '{print $5}')

echo ""
echo "--- 体积对比 ---"
echo ""
printf "| %-15s | %-12s | %-12s | %-10s |\n" "资源大小" "binary体积" "增长量" "膨胀率"
printf "| %-15s | %-12s | %-12s | %-10s |\n" "---" "---" "---" "---"
base_mb=$(echo "scale=2; $base_size / 1048576" | bc)
printf "| %-15s | %-12s | %-12s | %-10s |\n" "无嵌入(基准)" "${base_mb}MB" "-" "-"

for resource in resource_100k.bin resource_1m.bin resource_5m.bin resource_10m.bin; do
    res_size=$(ls -l $resource | awk '{print $5}')
    res_label=$(echo $resource | sed 's/resource_//' | sed 's/.bin//')
    
    cat > main_embed.go << GO
package main

import (
    _ "embed"
    "fmt"
)

//go:embed $resource
var data []byte

func main() { fmt.Printf("Embedded %d bytes\n", len(data)) }
GO
    
    CGO_ENABLED=0 go build -ldflags='-s -w' -o "embed_${res_label}" main_embed.go 2>/dev/null
    embed_size=$(ls -l "embed_${res_label}" | awk '{print $5}')
    growth=$((embed_size - base_size))
    growth_mb=$(echo "scale=2; $growth / 1048576" | bc)
    embed_mb=$(echo "scale=2; $embed_size / 1048576" | bc)
    ratio=$(echo "scale=1; $embed_size * 100 / $base_size" | bc)
    
    printf "| %-15s | %-12s | %-12s | %-10s |\n" "embed ${res_label}" "${embed_mb}MB" "+${growth_mb}MB" "${ratio}%"
done

echo ""
echo "--- 编译时间对比 ---"
echo ""

echo "无嵌入编译时间："
time_base=$( { TIMEFORMAT='%R'; time CGO_ENABLED=0 go build -o /dev/null main_base.go; } 2>&1 | tail -1)
echo "  ${time_base}s"

cat > main_embed_10m.go << 'GO'
package main

import (
    _ "embed"
    "fmt"
)

//go:embed resource_10m.bin
var data []byte

func main() { fmt.Printf("Embedded %d bytes\n", len(data)) }
GO

echo "嵌入10MB资源编译时间："
time_embed=$( { TIMEFORMAT='%R'; time CGO_ENABLED=0 go build -o /dev/null main_embed_10m.go; } 2>&1 | tail -1)
echo "  ${time_embed}s"

echo ""
echo "--- 部署灵活性分析 ---"
echo ""
echo "| 维度          | embed 方式   | 外置文件方式  |"
echo "|---------------|-------------|-------------|"
echo "| 分发复杂度     | 单文件分发   | binary+资源目录 |"
echo "| 资源更新       | 需重新编译   | 替换文件即可   |"
echo "| binary 大小   | 显著增大     | 保持最小      |"
echo "| 启动速度       | 无文件IO开销 | 需要读取文件   |"
echo "| 路径依赖       | 无          | 需处理路径    |"

# 清理
cd /
rm -rf "$TMPDIR"

echo ""
echo "=== 实验结束 ==="
