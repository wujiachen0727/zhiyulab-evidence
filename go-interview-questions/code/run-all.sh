#!/usr/bin/env bash
# 逐个运行 11 组实验（E1–E11），把真实输出落盘到 ../output/{实验名}/
# 用法：在 evidence/code/ 目录下执行 bash run-all.sh
set -u

cd "$(dirname "$0")"
OUT="../output"

mkdir -p "$OUT"/{make-vs-new,slice-append-alias,defer-semantics,defer-loop-alloc,map-concurrent,typed-nil-interface,goroutine-leak,channel-close,context-cancel,gomaxprocs,range-var}

run_experiment() {
    local name="$1"
    local pkg="$2"
    echo "── $name"
    go run "./$pkg" >"$OUT/$name/result.txt" 2>"$OUT/$name/stderr.txt"
    echo "   退出码 $?  ->  $OUT/$name/result.txt"
}

run_experiment make-vs-new         make-vs-new
run_experiment slice-append-alias  slice-append-alias
run_experiment defer-semantics     defer-semantics
run_experiment map-concurrent      map-concurrent
run_experiment typed-nil-interface typed-nil-interface
run_experiment goroutine-leak      goroutine-leak
run_experiment channel-close       channel-close
run_experiment context-cancel      context-cancel
run_experiment gomaxprocs          gomaxprocs

echo "── E4 第二半：循环内 defer 的分配方式（汇编对照）"
{
    echo "# 工具链：$(go version)"
    echo "# TEXT 行分隔函数；关注 CALL 的目标是 runtime.deferproc 还是只有 runtime.deferreturn"
    go build -gcflags=-S -o /dev/null ./defer-loop-alloc 2>&1 \
        | grep -E 'TEXT.*(nonLoop|inLoop)|runtime\.defer(proc|return)' | head -40
} >"$OUT/defer-loop-alloc/asm.txt" 2>&1
echo "   -> $OUT/defer-loop-alloc/asm.txt"

echo "── E3：同一份代码，三种语言版本（go.mod 的 go 行不同，工具链相同）"
# range-var-go1xx 是独立模块，必须进入各自目录执行
for v in 121 122 126; do
    (cd "range-var-go$v" && go run . ) >"$OUT/range-var/go$v.txt" 2>&1
    echo "   go.mod 声明 go 1.${v#1}.x -> $OUT/range-var/go$v.txt"
done

echo "── E11：库 API 的版本边界（sync.WaitGroup.Go，Go 1.25 新增）"
mkdir -p "$OUT/version-boundary"
(cd version-boundary/go124 && go run . ) >"$OUT/version-boundary/go124.txt" 2>&1
echo "   go 1.24 模块（预期编译失败） -> $OUT/version-boundary/go124.txt"
(cd version-boundary/go125 && go run . ) >"$OUT/version-boundary/go125.txt" 2>&1
echo "   go 1.25 模块（预期跑通）     -> $OUT/version-boundary/go125.txt"

echo
echo "全部完成。"
