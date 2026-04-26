// go/ast + go/types 版：精确检测"忽略 error 返回值"
// [实测 Go 1.26.2]
// 通过类型信息判断函数返回值中是否包含 error 类型
package main

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// TypesResult 记录一条类型检测结果
type TypesResult struct {
	File     string
	Line     int
	FuncName string
	Message  string
}

// RunTypesChecker 用 go/ast + go/types 检查给定包中忽略 error 的情况
func RunTypesChecker(pkgPattern string) ([]TypesResult, error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedName | packages.NeedFiles,
	}
	pkgs, loadErr := packages.Load(cfg, pkgPattern)
	if loadErr != nil {
		return nil, fmt.Errorf("加载包失败: %w", loadErr)
	}

	errorType := types.Universe.Lookup("error").Type()
	var results []TypesResult

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				return nil, fmt.Errorf("包错误: %v", e)
			}
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.ExprStmt:
					// 检测：函数调用结果被完全丢弃
					call, ok := node.X.(*ast.CallExpr)
					if !ok {
						return true
					}
					sig := resolveSignature(call, pkg)
					if sig == nil {
						return true
					}
					// 关键：只在返回值包含 error 类型时报告
					rets := sig.Results()
					for i := range rets.Len() {
						if types.Identical(rets.At(i).Type(), errorType) {
							pos := pkg.Fset.Position(call.Pos())
							name := exprName(call.Fun)
							results = append(results, TypesResult{
								File:     pos.Filename,
								Line:     pos.Line,
								FuncName: name,
								Message:  fmt.Sprintf("函数 %s 返回 error 但被丢弃", name),
							})
							break
						}
					}

			case *ast.AssignStmt:
				// 检测：用 _ 丢弃的是否是 error 类型
				// 策略：找到 RHS 的函数调用签名，匹配 LHS 中 _ 对应的返回值类型
				if len(node.Rhs) != 1 {
					return true
				}
				call, ok := node.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				sig := resolveSignature(call, pkg)
				if sig == nil {
					return true
				}
				rets := sig.Results()
				for i, lhs := range node.Lhs {
					ident, ok := lhs.(*ast.Ident)
					if !ok || ident.Name != "_" {
						continue
					}
					if i < rets.Len() && types.Identical(rets.At(i).Type(), errorType) {
						pos := pkg.Fset.Position(node.Pos())
						name := exprName(call.Fun)
						results = append(results, TypesResult{
							File:     pos.Filename,
							Line:     pos.Line,
							FuncName: name,
							Message:  fmt.Sprintf("_ 在位置 %d 丢弃了 %s 返回的 error", i+1, name),
						})
					}
				}
				}
				return true
			})
		}
	}

	return results, nil
}

// resolveSignature 从调用表达式解析函数签名
func resolveSignature(call *ast.CallExpr, pkg *packages.Package) *types.Signature {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		if obj := pkg.TypesInfo.ObjectOf(fn); obj != nil {
			if sig, ok := obj.Type().(*types.Signature); ok {
				return sig
			}
		}
	case *ast.SelectorExpr:
		if sel := pkg.TypesInfo.Selections[fn]; sel != nil {
			if sig, ok := sel.Type().(*types.Signature); ok {
				return sig
			}
		} else if obj := pkg.TypesInfo.ObjectOf(fn.Sel); obj != nil {
			if sig, ok := obj.Type().(*types.Signature); ok {
				return sig
			}
		}
	}
	return nil
}

// exprName 提取表达式名称
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

// assignRHSName 提取赋值语句右侧的函数名
func assignRHSName(assign *ast.AssignStmt) string {
	for _, rhs := range assign.Rhs {
		if call, ok := rhs.(*ast.CallExpr); ok {
			return exprName(call.Fun)
		}
	}
	return "<unknown>"
}
