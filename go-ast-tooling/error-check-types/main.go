// go/ast + go/types 版：精确检测"忽略 error 返回值"
// [实测 Go 1.26.2]
package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: %s <包路径>\n", os.Args[0])
		os.Exit(1)
	}

	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载失败: %v\n", err)
		os.Exit(1)
	}

	errorType := types.Universe.Lookup("error").Type()

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "包错误: %v\n", e)
			}
			continue
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				stmt, ok := n.(*ast.ExprStmt)
				if !ok {
					return true
				}
				call, ok := stmt.X.(*ast.CallExpr)
				if !ok {
					return true
				}

				// 用 TypesInfo 获取调用表达式的函数类型
				var sig *types.Signature
				switch fn := call.Fun.(type) {
				case *ast.Ident:
					if obj := pkg.TypesInfo.ObjectOf(fn); obj != nil {
						if f, ok := obj.Type().(*types.Signature); ok {
							sig = f
						}
					}
				case *ast.SelectorExpr:
					if sel := pkg.TypesInfo.Selections[fn]; sel != nil {
						if f, ok := sel.Type().(*types.Signature); ok {
							sig = f
						}
					} else if obj := pkg.TypesInfo.ObjectOf(fn.Sel); obj != nil {
						if f, ok := obj.Type().(*types.Signature); ok {
							sig = f
						}
					}
				}

				if sig == nil {
					return true
				}

				// 检查返回值是否包含 error 类型
				results := sig.Results()
				for i := 0; i < results.Len(); i++ {
					if types.Identical(results.At(i).Type(), errorType) {
						pos := pkg.Fset.Position(call.Pos())
						fmt.Printf("[Types] %s:%d  函数 %s 返回 error 但被丢弃\n",
							pos.Filename, pos.Line, exprName(call.Fun))
					}
				}
				return true
			})
		}
	}
}

func exprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprName(e.X), e.Sel.Name)
	default:
		return "<unknown>"
	}
}
