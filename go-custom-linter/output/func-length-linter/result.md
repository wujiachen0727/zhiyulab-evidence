# func-length-linter 实验输出
# [实测 Go 1.26.2] go/analysis + singlechecker 模式
# 实验时间：2026-04-26

## 环境
$ go version
go version go1.26.2 darwin/arm64

## 代码行数统计
$ wc -l analyzer.go main.go
      37 analyzer.go    # Analyzer 定义（含注释和空行）
       9 main.go        # singlechecker 入口
      46 total

# 纯逻辑行（去掉注释、空行、import、package）：
# analyzer.go: ~20 行（Analyzer 声明 + run 函数）
# main.go: 1 行（singlechecker.Main 调用）

## go test 输出（analysistest 框架）
$ go test -v ./...
=== RUN   TestAnalyzer
--- PASS: TestAnalyzer (0.40s)
PASS
ok  	funclength	0.651s

## 命令行运行输出
$ go run . ./testdata/src/example/
/Users/wujiachen/WriteCraft/articles/go-custom-linter/evidence/code/func-length-linter/testdata/src/example/example.go:15:1: 函数 TooLongFunc 有 84 行，超过上限 80 行
exit status 3

## 实验结论
1. 基于 go/analysis 框架写一个函数长度检查 linter，核心逻辑约 20 行
2. singlechecker.Main() 一行代码即可将 Analyzer 变为命令行工具
3. analysistest 框架通过 `// want` 注释实现声明式测试，零样板代码
4. 检测逻辑：计算 fn.Body.Rbrace 和 Lbrace 所在行号差值，超过阈值即报告
5. ShortFunc（5行）未触发 → TooLongFunc（84行）触发 → 检测正确
