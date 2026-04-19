// 函数行数检查器：用 go/ast 实现的最简代码分析工具
// [实测 Go 1.26.2]
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

const maxLines = 50

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: %s <go文件路径>\n", os.Args[0])
		os.Exit(1)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
		os.Exit(1)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		start := fset.Position(fn.Body.Lbrace)
		end := fset.Position(fn.Body.Rbrace)
		lines := end.Line - start.Line
		if lines > maxLines {
			fmt.Printf("⚠ %s:%d  函数 %s 有 %d 行（上限 %d）\n",
				start.Filename, start.Line, fn.Name.Name, lines, maxLines)
		}
		return true
	})
}
