// 函数长度检查器：基于 go/analysis 框架的最简 linter
// 检测函数体超过 80 行的函数并报告诊断信息
// [实测 Go 1.26.2]
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const maxLines = 80

var Analyzer = &analysis.Analyzer{
	Name: "funclength",
	Doc:  "检查函数体是否超过 80 行",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			start := pass.Fset.Position(fn.Body.Lbrace)
			end := pass.Fset.Position(fn.Body.Rbrace)
			lines := end.Line - start.Line
			if lines > maxLines {
				pass.Reportf(fn.Pos(), "函数 %s 有 %d 行，超过上限 %d 行", fn.Name.Name, lines, maxLines)
			}
			return true
		})
	}
	return nil, nil
}
