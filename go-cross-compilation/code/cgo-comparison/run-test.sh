#!/bin/bash
# E1: CGO=0 vs CGO=1 对比实验
# 测试同一个使用 sqlite 的 Go 项目在 CGO_ENABLED=0/1 下的编译差异
# 环境：macOS arm64, Go 1.22+

set -e

echo "=== E1: CGO=0 vs CGO=1 跨平台编译对比 ==="
echo ""
echo "测试环境："
go version
echo "OS: $(uname -s) / Arch: $(uname -m)"
echo ""

# 创建临时项目
TMPDIR=$(mktemp -d)
cd "$TMPDIR"

cat > go.mod << 'GOMOD'
module cgo-test
go 1.22
GOMOD

# 纯 Go 版本（使用 modernc sqlite，纯 Go 实现）
cat > main_pure.go << 'GOPURE'
package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

func main() {
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CGO: %v\n", os.Getenv("CGO_ENABLED"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s/%s", runtime.GOOS, runtime.GOARCH)
	})

	fmt.Println("Server ready (pure Go, no CGO)")
}
GOPURE

echo "--- 测试 1: 纯 Go HTTP 服务（无 CGO 依赖）---"
echo ""

# 编译多平台 binary（CGO=0）
echo "CGO_ENABLED=0 多平台编译："
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
for platform in "${platforms[@]}"; do
    os="${platform%/*}"
    arch="${platform#*/}"
    ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags='-s -w' -o "pure-${os}-${arch}${ext}" main_pure.go 2>&1
    size=$(ls -l "pure-${os}-${arch}${ext}" | awk '{print $5}')
    size_mb=$(echo "scale=2; $size / 1048576" | bc)
    echo "  ${os}/${arch}: ${size_mb} MB"
done

echo ""
echo "--- 测试 2: 尝试 CGO=1 交叉编译（预期失败）---"
echo ""

# 尝试 CGO=1 交叉编译到非当前平台
current_os=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$current_os" = "darwin" ]; then
    target_os="linux"
else
    target_os="darwin"
fi

echo "当前平台: ${current_os}/arm64"
echo "目标平台: ${target_os}/amd64"
echo ""

# CGO=1 本地编译
echo "CGO_ENABLED=1 本地编译（当前平台）："
CGO_ENABLED=1 go build -ldflags='-s -w' -o "cgo-local" main_pure.go 2>&1
local_size=$(ls -l "cgo-local" | awk '{print $5}')
local_size_mb=$(echo "scale=2; $local_size / 1048576" | bc)
echo "  本地 binary: ${local_size_mb} MB"
echo ""

# CGO=1 交叉编译（会失败，除非有交叉编译工具链）
echo "CGO_ENABLED=1 交叉编译到 ${target_os}/amd64："
if CGO_ENABLED=1 GOOS=$target_os GOARCH=amd64 go build -ldflags='-s -w' -o "cgo-cross" main_pure.go 2>&1; then
    cross_size=$(ls -l "cgo-cross" | awk '{print $5}')
    cross_size_mb=$(echo "scale=2; $cross_size / 1048576" | bc)
    echo "  交叉编译 binary: ${cross_size_mb} MB"
else
    echo "  ❌ 编译失败：CGO_ENABLED=1 需要目标平台的 C 编译工具链"
    echo "  这就是 CGO 跨平台编译的核心痛点"
fi

echo ""
echo "--- 测试 3: 编译速度对比 ---"
echo ""

echo "CGO_ENABLED=0 编译时间（本地平台）："
time_cgo0=$(TIMEFORMAT='%R'; { time CGO_ENABLED=0 go build -o /dev/null main_pure.go; } 2>&1)
echo "  ${time_cgo0}s"

echo "CGO_ENABLED=1 编译时间（本地平台）："
time_cgo1=$(TIMEFORMAT='%R'; { time CGO_ENABLED=1 go build -o /dev/null main_pure.go; } 2>&1)
echo "  ${time_cgo1}s"

echo ""
echo "--- 测试 4: 依赖分析 ---"
echo ""

echo "CGO=0 binary 动态链接依赖："
CGO_ENABLED=0 go build -o "static-binary" main_pure.go 2>&1
if command -v otool &> /dev/null; then
    otool -L "static-binary" 2>/dev/null || echo "  （无动态链接依赖 — 纯静态）"
elif command -v ldd &> /dev/null; then
    ldd "static-binary" 2>/dev/null || echo "  （无动态链接依赖 — 纯静态）"
fi

echo ""
echo "CGO=1 binary 动态链接依赖："
CGO_ENABLED=1 go build -o "dynamic-binary" main_pure.go 2>&1
if command -v otool &> /dev/null; then
    otool -L "dynamic-binary" 2>/dev/null || echo "  无法分析"
elif command -v ldd &> /dev/null; then
    ldd "dynamic-binary" 2>/dev/null || echo "  无法分析"
fi

echo ""
echo "--- 测试 5: 支持的平台数量 ---"
echo ""
echo "Go 官方支持的 GOOS/GOARCH 组合数："
go tool dist list | wc -l | xargs echo "  "
echo ""
echo "其中 CGO_ENABLED=0 可直接交叉编译的组合数：全部（$(go tool dist list | wc -l | xargs) 个）"
echo "CGO_ENABLED=1 可直接交叉编译的组合数：仅当前平台（1 个，除非安装交叉工具链）"

# 清理
cd /
rm -rf "$TMPDIR"

echo ""
echo "=== 实验结束 ==="
