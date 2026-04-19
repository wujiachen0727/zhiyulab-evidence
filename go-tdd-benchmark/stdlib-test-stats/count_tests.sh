#!/bin/bash
# 统计 Go 标准库中 Test/Bench/Sub 的使用比例
GOROOT=$(go env GOROOT)
echo "Go 版本: $(go version)"
echo "GOROOT: $GOROOT"
echo ""

# 统计所有 *_test.go 文件
test_files=$(find "$GOROOT/src" -name "*_test.go" 2>/dev/null | wc -l)
echo "标准库 *_test.go 文件数: $test_files"

# 统计 func TestXxx
func_test=$(grep -r "^func Test" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "func TestXxx 数量: $func_test"

# 统计 func BenchmarkXxx
func_bench=$(grep -r "^func Benchmark" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "func BenchmarkXxx 数量: $func_bench"

# 统计 func FuzzXxx
func_fuzz=$(grep -r "^func Fuzz" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "func FuzzXxx 数量: $func_fuzz"

# 统计 t.Run 子测试
t_run=$(grep -r "t\.Run(" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "t.Run() 子测试调用数: $t_run"

# 统计 b.Run 子测试
b_run=$(grep -r "b\.Run(" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "b.Run() 子测试调用数: $b_run"

# 统计 t.Helper()
t_helper=$(grep -r "t\.Helper()" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "t.Helper() 调用数: $t_helper"

# 统计 testify 使用
testify=$(grep -r "testify" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "testify 引用数: $testify"

# 统计第三方 assert 使用
third_assert=$(grep -rE "assert\.(Equal|True|False|Nil|NotNil|Error|NoError)" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "第三方 assert 风格调用数: $third_assert"

# 统计 if + Errorf/Error/Fatal 模式
if_errorf=$(grep -rE "if .+ \!= .+ \{" "$GOROOT/src" --include="*_test.go" 2>/dev/null | wc -l)
echo "if 条件判断模式数: $if_errorf"

echo ""
echo "=== 比例分析 ==="
total_top_level=$((func_test + func_bench + func_fuzz))
if [ $total_top_level -gt 0 ]; then
  test_pct=$(echo "scale=1; $func_test * 100 / $total_top_level" | bc)
  bench_pct=$(echo "scale=1; $func_bench * 100 / $total_top_level" | bc)
  fuzz_pct=$(echo "scale=1; $func_fuzz * 100 / $total_top_level" | bc)
  echo "Test 占比: ${test_pct}%"
  echo "Benchmark 占比: ${bench_pct}%"
  echo "Fuzz 占比: ${fuzz_pct}%"
fi

echo ""
echo "Test : Benchmark 比例 = $func_test : $func_bench"
if [ $func_bench -gt 0 ]; then
  ratio=$(echo "scale=1; $func_test / $func_bench" | bc)
  echo "Test 是 Benchmark 的 ${ratio} 倍"
fi
