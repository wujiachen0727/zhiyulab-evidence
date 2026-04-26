// main.go — 架构分层 linter 实验: go/ast vs go/types 方案对比
// [实测 Go 1.26.2]
//
// 实验目的：展示为什么实现"禁止 Handler 层直接调用 Repository 层"的
// 架构约束检查需要 go/types，纯 go/ast 方案存在根本性局限。

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  架构分层 Linter 实验: go/ast vs go/types                  ║")
	fmt.Println("║  规则: Handler 层禁止直接调用 Repository 层                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 定位 testdata 目录
	testdataDir, err := findTestdataDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	handlerDir := filepath.Join(testdataDir, "handler")
	fmt.Printf("测试目录: %s\n", testdataDir)
	fmt.Println()

	// 列出 handler 中的测试场景
	fmt.Println("━━━ 测试场景 ━━━")
	fmt.Println("  1. HandleDeleteUser_ViolationDirect  — 直接调用 repo.UserRepo{}（带 import alias）")
	fmt.Println("  2. HandleGetUser_ViolationField      — 通过 struct 字段 h.repoObj 调用")
	fmt.Println("  3. HandleGetUser_ViolationInterface   — 通过接口字段 h.repoIntf 调用（go/ast 盲区）")
	fmt.Println("  4. HandleGetUser_Compliant            — 合规: 通过 service 层调用")
	fmt.Println()

	// ===== 方案一：纯 go/ast =====
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	astResults, err := runASTChecker(handlerDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AST 检测失败: %v\n", err)
		os.Exit(1)
	}
	printASTResults(astResults)

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// ===== 方案二：go/ast + go/types =====
	typesResults, err := runTypesChecker(testdataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Types 检测失败: %v\n", err)
		os.Exit(1)
	}
	printTypesResults(typesResults)

	// ===== 对比总结 =====
	fmt.Println()
	fmt.Println("━━━ 对比总结 ━━━")
	fmt.Println()

	astCount := len(astResults)
	typesCount := len(typesResults)

	fmt.Printf("  实际违规调用: 3 处（案例 1、2、3）\n")
	fmt.Printf("  go/ast  检出: %d 处\n", astCount)
	fmt.Printf("  go/types 检出: %d 处\n", typesCount)
	fmt.Println()

	if typesCount > astCount {
		fmt.Printf("  go/types 多检出 %d 处——这些是 go/ast 的盲区:\n", typesCount-astCount)
		// 找出 types 有但 ast 没有的
		astLines := make(map[int]bool)
		for _, r := range astResults {
			astLines[r.Line] = true
		}
		for _, r := range typesResults {
			if !astLines[r.Line] {
				fmt.Printf("    - %s:%d %s [%s]\n", r.File, r.Line, r.FuncName, r.Category)
			}
		}
	}

	fmt.Println()
	fmt.Println("━━━ 结论 ━━━")
	fmt.Println()
	fmt.Println("  纯 go/ast 只能做字符串级匹配（import alias + 变量名启发式），")
	fmt.Println("  存在漏检（接口调用）和误报（名字碰巧匹配）的双重风险。")
	fmt.Println()
	fmt.Println("  go/types 通过类型系统追踪每个调用的真实目标包路径，")
	fmt.Println("  不受 alias、接口、中间变量的影响——这才是 production-ready 的方案。")
}

// findTestdataDir 查找 testdata 目录
func findTestdataDir() (string, error) {
	// 尝试从可执行文件位置推断
	exec, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exec)
		candidate := filepath.Join(dir, "testdata")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 尝试从当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("无法获取工作目录: %w", err)
	}

	// 直接在当前目录的 testdata
	candidate := filepath.Join(wd, "testdata")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// 向上查找包含 testdata 的 architecture-linter 目录
	dir := wd
	for {
		if strings.HasSuffix(dir, "architecture-linter") {
			candidate := filepath.Join(dir, "testdata")
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("找不到 testdata 目录（当前目录: %s）", wd)
}
