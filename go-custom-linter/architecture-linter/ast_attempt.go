// ast_attempt.go — 纯 go/ast 方案尝试检测 handler→repository 越级调用
// [实测 Go 1.26.2]
//
// 局限性：
//   1. 只能匹配 import 路径中包含 "repository" 的包名/alias
//   2. import alias 不同时（如 repo "xxx/repository"），需要额外追踪 alias 映射
//   3. 通过接口变量调用 repository 方法时完全无法检测——ast 看不到类型信息
//   4. 通过中间变量传递 repository 对象时也无法追踪

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ASTResult 纯 AST 检测结果
type ASTResult struct {
	File     string
	Line     int
	Col      int
	FuncName string
	Receiver string
	Detail   string
}

// runASTChecker 用纯 go/ast 检测 handler 包中对 repository 包的直接调用
func runASTChecker(handlerDir string) ([]ASTResult, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, handlerDir, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("解析目录失败: %w", err)
	}

	var results []ASTResult

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			// 第一步：收集 import alias → 包路径 的映射
			aliasToPath := make(map[string]string)
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				var alias string
				if imp.Name != nil {
					alias = imp.Name.Name
				} else {
					// 默认 alias = 包路径最后一段
					parts := strings.Split(path, "/")
					alias = parts[len(parts)-1]
				}
				aliasToPath[alias] = path
			}

			// 第二步：找出哪些 alias 指向 repository 包
			repoAliases := make(map[string]bool)
			for alias, path := range aliasToPath {
				if strings.Contains(path, "repository") {
					repoAliases[alias] = true
				}
			}

			if len(repoAliases) == 0 {
				continue // 没有导入 repository 包
			}

			// 第三步：遍历 AST，查找 selector 表达式中使用了 repository alias 的调用
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// 检查直接函数调用: repo.XXX()
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if repoAliases[ident.Name] {
							pos := fset.Position(call.Pos())
							results = append(results, ASTResult{
								File:     filepath.Base(filename),
								Line:     pos.Line,
								Col:      pos.Column,
								FuncName: sel.Sel.Name,
								Receiver: ident.Name,
								Detail:   fmt.Sprintf("检测到 %s.%s() 调用（alias 匹配）", ident.Name, sel.Sel.Name),
							})
						}
					}
					// ⚠️ 局限：如果 sel.X 不是简单 Ident（比如 h.repoObj.Method()），
					// 这里就需要递归解析——但 AST 无法确定 h.repoObj 的类型
				}

				return true
			})

			// 第四步：尝试检测 obj.Method() 模式（字段调用）
			// AST 能看到 h.repoObj.FindByID()，但无法确定 repoObj 的类型
			// 只能做启发式匹配：字段名包含 "repo" 的视为可疑
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					// 检查 X 是否是 a.b 形式（h.repoObj）
					if innerSel, ok := sel.X.(*ast.SelectorExpr); ok {
						fieldName := innerSel.Sel.Name
						if strings.Contains(strings.ToLower(fieldName), "repo") &&
							!strings.Contains(strings.ToLower(fieldName), "repointf") {
							// ⚠️ 启发式：只能靠字段名猜测
							pos := fset.Position(call.Pos())
							results = append(results, ASTResult{
								File:     filepath.Base(filename),
								Line:     pos.Line,
								Col:      pos.Column,
								FuncName: sel.Sel.Name,
								Receiver: fieldName,
								Detail:   fmt.Sprintf("启发式检测: 字段 %s 名称含 'repo'（不确定是否真的是 repository 调用）", fieldName),
							})
						}
						// ⚠️ 对 h.repoIntf.FindByID() 完全无能为力
						// 因为字段名 "repoIntf" 不一定包含 "repo"，
						// 即使包含也无法确定其类型是否来自 repository 包
					}
				}
				return true
			})
		}
	}

	return results, nil
}

// printASTResults 打印纯 AST 检测结果
func printASTResults(results []ASTResult) {
	fmt.Println("=== 纯 go/ast 方案检测结果 ===")
	fmt.Println()
	if len(results) == 0 {
		fmt.Println("  未检测到违规调用")
		return
	}
	for i, r := range results {
		fmt.Printf("  [%d] %s:%d:%d\n", i+1, r.File, r.Line, r.Col)
		fmt.Printf("      %s\n", r.Detail)
	}
	fmt.Println()
	fmt.Println("⚠️  go/ast 的局限性:")
	fmt.Println("  - 依赖字符串匹配 import alias/字段名，alias 改变就失效")
	fmt.Println("  - 无法检测通过接口变量（如 h.repoIntf）的 repository 调用")
	fmt.Println("  - 启发式匹配（靠字段名猜）既有漏检又有误报")
}
