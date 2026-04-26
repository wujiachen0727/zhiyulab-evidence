// error-check-compare：对比 go/ast 和 go/types 检测"忽略 error"的精度差异
// [实测 Go 1.26.2]
//
// 实验设计：
//   - testdata/target.go 包含 10 种场景
//     - 4 种真阳性（确实忽略了危险的 error）
//     - 3 种 AST 假阳性（不涉及 error，但 AST 会误报）
//     - 1 种 AST+Types 共同假阳性（fmt.Println 的 error 工程上可忽略）
//     - 2 种真阴性（正确处理或无 error）
//   - ast_checker.go 用纯 go/ast 做语法模式匹配
//   - types_checker.go 用 go/ast + go/types 做类型感知检查
//   - 本文件运行两个 checker 并输出对比报告
package main

import (
	"fmt"
	"os"
	"strings"
)

// 场景标注：行号 → 分类
type scenarioClass int

const (
	truePositive  scenarioClass = iota // 确实忽略了危险的 error
	falsePositive                      // 不涉及 error 或工程上可忽略
	trueNegative                       // 正确处理，不应报告
)

type scenario struct {
	line  int
	class scenarioClass
	desc  string
}

// 按 testdata/target.go 的实际行号标注
var scenarios = []scenario{
	// 真阳性：确实忽略了危险的 error
	{16, truePositive, "os.Remove() — 完全丢弃 error"},
	{21, truePositive, "f, _ := os.Open() — 用 _ 丢弃 error"},
	{27, truePositive, "v, _ := strconv.Atoi() — 用 _ 丢弃 error"},
	{33, truePositive, "os.MkdirAll() — 完全丢弃 error"},

	// AST 会误报的：不涉及 error
	{48, falsePositive, "strings.Contains() — 返回 bool，不返回 error"},
	{58, falsePositive, "_ = f — 丢弃的是 *os.File，不是 error"},
	{64, falsePositive, "doNothing() — 函数无返回值"},

	// fmt.Println 特殊：技术上返回 error，但工程上可忽略
	{42, falsePositive, "fmt.Println() — 技术上返回 error，工程上可忽略"},
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  error 忽略检测精度对比：go/ast vs go/types")
	fmt.Println("  [实测 Go 1.26.2]")
	fmt.Println(strings.Repeat("=", 70))

	// ---- 运行纯 AST checker ----
	fmt.Println("\n--- 纯 go/ast 检测结果 ---")
	astResults, err := RunASTChecker("testdata/target.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "AST checker 失败: %v\n", err)
		os.Exit(1)
	}
	for i, r := range astResults {
		fmt.Printf("  [AST #%d] 第%d行: %s\n", i+1, r.Line, r.Message)
	}
	fmt.Printf("\n  AST 检测总数: %d\n", len(astResults))

	// ---- 运行 Types checker ----
	fmt.Println("\n--- go/ast + go/types 检测结果 ---")
	typesResults, err := RunTypesChecker("./testdata")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Types checker 失败: %v\n", err)
		os.Exit(1)
	}
	for i, r := range typesResults {
		fmt.Printf("  [Types #%d] 第%d行: %s — %s\n", i+1, r.Line, r.FuncName, r.Message)
	}
	fmt.Printf("\n  Types 检测总数: %d\n", len(typesResults))

	// ---- 建立行号索引 ----
	scenarioMap := make(map[int]scenario)
	for _, s := range scenarios {
		scenarioMap[s.line] = s
	}

	// ---- 统计 AST ----
	astTP, astFP, astNoise := 0, 0, 0
	for _, r := range astResults {
		s, known := scenarioMap[r.Line]
		if !known {
			// 不在预期场景中的检测（如 _ = v 这种中间变量）
			astNoise++
			continue
		}
		if s.class == truePositive {
			astTP++
		} else {
			astFP++
		}
	}

	// ---- 统计 Types ----
	typesTP, typesFP := 0, 0
	for _, r := range typesResults {
		s, known := scenarioMap[r.Line]
		if !known {
			typesFP++
			continue
		}
		if s.class == truePositive {
			typesTP++
		} else {
			typesFP++
		}
	}

	// ---- 输出对比表格 ----
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  精度对比分析")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Println("\n  预期真阳性（确实忽略了危险 error 的场景）:")
	for _, s := range scenarios {
		if s.class == truePositive {
			fmt.Printf("    第%d行: %s\n", s.line, s.desc)
		}
	}
	fmt.Printf("    共 4 个\n")

	astTotal := astTP + astFP + astNoise
	typesTotal := typesTP + typesFP

	fmt.Println("\n  ┌───────────────────────┬──────────┬─────────────┐")
	fmt.Println("  │       指标            │  go/ast  │ go/ast+types│")
	fmt.Println("  ├───────────────────────┼──────────┼─────────────┤")
	fmt.Printf("  │ 检测总数              │    %2d    │      %2d     │\n", astTotal, typesTotal)
	fmt.Printf("  │ 真阳性（命中）        │    %2d    │      %2d     │\n", astTP, typesTP)
	fmt.Printf("  │ 假阳性（噪音）        │    %2d    │      %2d     │\n", astFP+astNoise, typesFP)
	astPrec := safePrecision(astTP, astTotal)
	typesPrec := safePrecision(typesTP, typesTotal)
	fmt.Printf("  │ 精确率                │  %5.1f%%  │    %5.1f%%   │\n", astPrec, typesPrec)
	fmt.Printf("  │ 召回率（/4 真阳性）   │  %5.1f%%  │    %5.1f%%   │\n",
		float64(astTP)/4*100, float64(typesTP)/4*100)
	fmt.Println("  └───────────────────────┴──────────┴─────────────┘")

	// ---- AST 假阳性详情 ----
	fmt.Println("\n  AST 假阳性/噪音详情:")
	for _, r := range astResults {
		s, known := scenarioMap[r.Line]
		if !known {
			fmt.Printf("    第%d行: [噪音] %s\n", r.Line, r.Message)
		} else if s.class == falsePositive {
			fmt.Printf("    第%d行: [假阳性] %s\n", r.Line, s.desc)
		}
	}

	// ---- Types 假阳性详情 ----
	if typesFP > 0 {
		fmt.Println("\n  Types 假阳性详情:")
		for _, r := range typesResults {
			s, known := scenarioMap[r.Line]
			if !known || s.class != truePositive {
				fmt.Printf("    第%d行: %s — %s\n", r.Line, r.FuncName, r.Message)
			}
		}
	}

	// ---- 结论 ----
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("  结论:")
	fmt.Println()
	fmt.Println("  纯 go/ast：")
	fmt.Println("    - 只能做语法模式匹配（看到 _ 就报、看到 ExprStmt 就报）")
	fmt.Println("    - 无法区分被丢弃的值是 error、bool、*File 还是无返回值")
	fmt.Printf("    - 精确率仅 %.0f%%，大量噪音淹没真正的问题\n", astPrec)
	fmt.Println()
	fmt.Println("  go/ast + go/types：")
	fmt.Println("    - 通过 types.Signature 精确判断函数返回值中是否包含 error")
	fmt.Println("    - 跳过 strings.Contains (bool)、doNothing() (无返回值) 等")
	fmt.Printf("    - 精确率 %.0f%%，只报告真正涉及 error 的场景\n", typesPrec)
	fmt.Println()
	fmt.Println("  这就是为什么生产级 linter 都基于 go/types 而非纯 go/ast：")
	fmt.Println("  没有类型信息，语法检查器无法区分 os.Remove()（危险）和 doNothing()（无害）。")
	fmt.Println(strings.Repeat("-", 70))
}

func safePrecision(tp, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(tp) / float64(total) * 100
}
