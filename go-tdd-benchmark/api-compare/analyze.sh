#!/bin/bash
# 统计 Go testing 包的公共 API 面积
echo "=== Go testing 包 API 统计 ==="
GOROOT=$(go env GOROOT)

# 统计导出类型
types=$(grep -rE "^type [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l)
echo "导出类型数: $types"

# 统计导出函数
funcs=$(grep -rE "^func [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l)
echo "导出函数数: $funcs"

# 统计导出变量/常量
vars=$(grep -rE "^(var|const) [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l)
echo "导出变量/常量数: $vars"

# 列出核心类型
echo ""
echo "=== 核心导出类型 ==="
grep -rE "^type [A-Z]\w+ " "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | sed 's/.*src\/testing\///' | sort

# 列出 T 的导出方法
echo ""
echo "=== testing.T 的导出方法 ==="
grep -rE "^func \(.*\*?T\) [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | sed 's/.*func (.*\*?T) //' | sed 's/ {.*/()/' | sort

# 列出 B 的导出方法
echo ""
echo "=== testing.B 的导出方法 ==="
grep -rE "^func \(.*\*?B\) [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | sed 's/.*func (.*\*?B) //' | sed 's/ {.*/()/' | sort

# 列出 F 的导出方法
echo ""
echo "=== testing.F 的导出方法 ==="
grep -rE "^func \(.*\*?F\) [A-Z]" "$GOROOT/src/testing/" --include="*.go" 2>/dev/null | grep -v "_test.go" | sed 's/.*func (.*\*?F) //' | sed 's/ {.*/()/' | sort

# 统计文件数和总行数
echo ""
echo "=== 包规模 ==="
go_files=$(find "$GOROOT/src/testing/" -name "*.go" ! -name "*_test.go" 2>/dev/null | wc -l)
go_lines=$(find "$GOROOT/src/testing/" -name "*.go" ! -name "*_test.go" -exec cat {} + 2>/dev/null | wc -l)
echo "非测试 .go 文件数: $go_files"
echo "总代码行数: $go_lines"
