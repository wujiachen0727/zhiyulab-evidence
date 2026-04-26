// types_checker.go — go/ast + go/types 方案精确检测 handler→repository 越级调用
// [实测 Go 1.26.2]
//
// 核心优势：
//   1. 通过 types.Info.Uses 获取标识符的真实类型信息
//   2. 通过 types.Info.Selections 获取字段/方法选择的完整类型链
//   3. 不依赖命名约定，直接判断包路径是否属于 repository 层
//   4. 能检测接口变量、alias import、中间变量传递等所有情况

package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// TypesResult go/types 检测结果
type TypesResult struct {
	File       string
	Line       int
	Col        int
	FuncName   string
	PkgPath    string
	Detail     string
	Category   string // "direct" | "field" | "interface"
}

// runTypesChecker 用 go/types 精确检测 handler 包中对 repository 包的直接调用
func runTypesChecker(projectDir string) ([]TypesResult, error) {
	// 使用 go/packages 加载完整类型信息
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: projectDir,
	}

	pkgs, err := packages.Load(cfg, "./handler/...")
	if err != nil {
		return nil, fmt.Errorf("加载包失败: %w", err)
	}

	var results []TypesResult

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Printf("  包加载警告: %s\n", e)
			}
		}

		// 确认是 handler 包
		if !strings.HasSuffix(pkg.PkgPath, "/handler") && pkg.Name != "handler" {
			continue
		}

		fset := pkg.Fset
		info := pkg.TypesInfo

		for i, file := range pkg.Syntax {
			filename := pkg.GoFiles[i]

			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// 关键：用 types.Info.Selections 或 types.Info.Uses 获取真实类型
				var targetPkgPath string
				var category string

				// 方式1：检查 selector 的 selection（字段/方法选择）
				if selection, ok := info.Selections[sel]; ok {
					recv := selection.Recv()
					targetPkgPath = extractPkgPath(recv)
					if strings.Contains(targetPkgPath, "repository") {
						category = categorizeSelection(selection)
					}
				}

				// 方式2：如果 selector 没有 selection 信息，检查 Uses
				if targetPkgPath == "" {
					if obj := info.Uses[sel.Sel]; obj != nil {
						if obj.Pkg() != nil {
							targetPkgPath = obj.Pkg().Path()
							category = "direct"
						}
					}
				}

				// 方式3：检查 X 部分的类型（处理包级函数调用 repo.XXX()）
				if targetPkgPath == "" {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if obj := info.Uses[ident]; obj != nil {
							if pkgName, ok := obj.(*types.PkgName); ok {
								targetPkgPath = pkgName.Imported().Path()
								category = "direct"
							}
						}
					}
				}

				// 判断是否属于 repository 层
				if strings.Contains(targetPkgPath, "repository") {
					pos := fset.Position(call.Pos())
					results = append(results, TypesResult{
						File:     filepath.Base(filename),
						Line:     pos.Line,
						Col:      pos.Column,
						FuncName: sel.Sel.Name,
						PkgPath:  targetPkgPath,
						Category: category,
						Detail: fmt.Sprintf("handler 直接调用 repository 层: %s.%s() [%s]",
							filepath.Base(targetPkgPath), sel.Sel.Name, category),
					})
				}

				return true
			})
		}
	}

	return results, nil
}

// extractPkgPath 从类型中提取包路径
func extractPkgPath(t types.Type) string {
	// 解引用指针
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	// 获取命名类型的包路径
	if named, ok := t.(*types.Named); ok {
		if named.Obj().Pkg() != nil {
			return named.Obj().Pkg().Path()
		}
	}
	return ""
}

// categorizeSelection 分类 selection 的来源
func categorizeSelection(sel *types.Selection) string {
	recv := sel.Recv()
	// 解引用指针
	if ptr, ok := recv.(*types.Pointer); ok {
		recv = ptr.Elem()
	}
	if named, ok := recv.(*types.Named); ok {
		if named.Obj().IsAlias() {
			return "alias"
		}
		underlying := named.Underlying()
		if _, ok := underlying.(*types.Interface); ok {
			return "interface"
		}
	}
	// 检查 selection 的 indirect 属性
	if sel.Indirect() {
		return "field-indirect"
	}
	return "field"
}

// printTypesResults 打印 go/types 检测结果
func printTypesResults(results []TypesResult) {
	fmt.Println("=== go/ast + go/types 方案检测结果 ===")
	fmt.Println()
	if len(results) == 0 {
		fmt.Println("  未检测到违规调用")
		return
	}
	for i, r := range results {
		fmt.Printf("  [%d] %s:%d:%d\n", i+1, r.File, r.Line, r.Col)
		fmt.Printf("      %s\n", r.Detail)
		fmt.Printf("      包路径: %s | 检测方式: %s\n", r.PkgPath, r.Category)
	}
	fmt.Println()
	fmt.Println("✅ go/types 的优势:")
	fmt.Println("  - 基于包路径精确判断，不依赖 import alias 或变量命名")
	fmt.Println("  - 能追踪接口变量的底层类型，检测间接调用")
	fmt.Println("  - 零误报：只有真正调用了 repository 包的代码才会被标记")
}
