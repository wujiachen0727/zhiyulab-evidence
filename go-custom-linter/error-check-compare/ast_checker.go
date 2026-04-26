// 纯 go/ast 版：检测"忽略 error 返回值"
// [实测 Go 1.26.2]
// 局限性：只能做语法层面的模式匹配，无法知道函数返回值类型
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// ASTResult 记录一条 AST 检测结果
type ASTResult struct {
	File    string
	Line    int
	Message string
}

// RunASTChecker 用纯 go/ast 检查给定文件中忽略 error 的情况
func RunASTChecker(filename string) ([]ASTResult, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}

	var results []ASTResult

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			// 检测模式1：赋值语句中左侧有 _
			// 问题：无法区分被丢弃的是 error 还是其他类型
			for _, lhs := range node.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if ok && ident.Name == "_" {
					pos := fset.Position(node.Pos())
					results = append(results, ASTResult{
						File:    pos.Filename,
						Line:    pos.Line,
						Message: fmt.Sprintf("发现 _ 赋值（可能忽略 error）: %s", formatAssign(node, fset)),
					})
					break // 同一语句只报一次
				}
			}

		case *ast.ExprStmt:
			// 检测模式2：函数调用结果被完全丢弃
			// 问题：不知道函数是否有返回值，更不知道是否返回 error
			call, ok := node.X.(*ast.CallExpr)
			if !ok {
				return true
			}
			pos := fset.Position(call.Pos())
			results = append(results, ASTResult{
				File:    pos.Filename,
				Line:    pos.Line,
				Message: fmt.Sprintf("函数调用返回值被丢弃: %s()", callName(call.Fun)),
			})
		}
		return true
	})

	return results, nil
}

// callName 提取函数调用名（如 fmt.Println、os.Remove）
func callName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", callName(e.X), e.Sel.Name)
	default:
		return "<unknown>"
	}
}

// formatAssign 格式化赋值语句
func formatAssign(assign *ast.AssignStmt, fset *token.FileSet) string {
	lhs := ""
	for i, l := range assign.Lhs {
		if i > 0 {
			lhs += ", "
		}
		if ident, ok := l.(*ast.Ident); ok {
			lhs += ident.Name
		} else {
			lhs += "?"
		}
	}
	rhs := ""
	for i, r := range assign.Rhs {
		if i > 0 {
			rhs += ", "
		}
		if call, ok := r.(*ast.CallExpr); ok {
			rhs += callName(call.Fun) + "()"
		} else {
			rhs += "..."
		}
	}
	op := assign.Tok.String()
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}
