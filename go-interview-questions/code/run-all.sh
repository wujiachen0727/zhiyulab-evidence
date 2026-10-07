#!/usr/bin/env bash
# 逐个运行 13 组实验，把真实输出落盘到 ../output/{实验名}/
#
# 用法：在 code/ 目录下执行 bash run-all.sh
#
# 说明：每个实验是独立 module，必须进入各自目录执行 go run .
#       range-var-go121 / go122 / go126 三份源码完全相同，唯一差别是 go.mod 的 go 行，
#       用来演示 for range 循环变量语义由声明的语言版本决定（而不是工具链版本）。
set -u

cd "$(dirname "$0")"
OUT="../output"

mkdir -p "$OUT"/{make-vs-new,slice-append-alias,defer-semantics,defer-loop-alloc,map-concurrent,typed-nil-interface,goroutine-leak,channel-close,context-cancel,gomaxprocs,range-var}

run() {
    local name="$1"
    echo "── $name"
    (cd "$name" && go run .) >"$OUT/$name/result.txt" 2>"$OUT/$name/stderr.txt"
    echo "   退出码 $?  ->  $OUT/$name/result.txt"
}

run make-vs-new
run slice-append-alias
run defer-semantics
run map-concurrent
run typed-nil-interface
run goroutine-leak
run channel-close
run context-cancel
run gomaxprocs

echo "── defer-loop-alloc：循环内 defer 的分配方式（汇编对照）"
{
    echo "# 工具链：$(go version)"
    echo "# TEXT 行分隔函数；关注 CALL 的目标是 runtime.deferproc 还是只有 runtime.deferreturn"
    (cd defer-loop-alloc && go build -gcflags=-S -o /dev/null . 2>&1) \
        | grep -E 'TEXT.*(nonLoop|inLoop)|runtime\.defer(proc|return)' | head -40
} >"$OUT/defer-loop-alloc/asm.txt" 2>&1
echo "   -> $OUT/defer-loop-alloc/asm.txt"

echo "── for range 循环变量：三种语言版本对照"
for v in 121 122 126; do
    (cd "range-var-go$v" && go run .) >"$OUT/range-var/go$v.txt" 2>&1
    echo "   go.mod 声明 go 1.${v#1}.x -> $OUT/range-var/go$v.txt"
done

echo
echo "全部完成。"
