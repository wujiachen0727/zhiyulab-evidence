// 纯 go/ast 版：检测"忽略 error 返回值"
// [实测 Go 1.26.2]
// 局限性：只能做文本匹配，无法知道函数返回值类型
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

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

	// 纯 AST 方式：只能检测赋值语句中左侧有 _ 的情况
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if ok && ident.Name == "_" {
				pos := fset.Position(assign.Pos())
				fmt.Printf("[AST] %s:%d  发现 _ 赋值，可能忽略了 error\n",
					pos.Filename, pos.Line)
			}
		}
		return true
	})

	// 纯 AST 方式也无法检测：直接丢弃返回值（不赋值）
	// 例如: os.Remove("file.txt")  ← 这里 error 被完全丢弃，AST 层面看不出问题
	ast.Inspect(f, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		pos := fset.Position(call.Pos())
		// 问题在这里：我们不知道这个函数是否有返回值，更不知道返回值是不是 error
		// 只能"猜"——这就是纯 AST 的天花板
		fmt.Printf("[AST] %s:%d  函数调用的返回值被丢弃（但我们不知道它是否返回 error）\n",
			pos.Filename, pos.Line)
		return true
	})
}
