package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var Analyzer = &analysis.Analyzer{
	Name: "funclen",
	Doc:  "检查函数行数是否超过上限",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Body == nil {
				return true
			}
			start := pass.Fset.Position(fn.Body.Lbrace)
			end := pass.Fset.Position(fn.Body.Rbrace)
			if end.Line-start.Line > 50 {
				pass.Reportf(fn.Pos(), "函数 %s 有 %d 行（上限 50）",
					fn.Name.Name, end.Line-start.Line)
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	singlechecker.Main(Analyzer)
}
