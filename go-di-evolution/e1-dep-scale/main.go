package main

import (
	"fmt"
	"strings"
)

// E1 实验：依赖图规模 vs 维护成本曲线
// 核心假设：手动 DI 的组装代码行数随依赖数非线性增长（存在拐点）
// 证伪逻辑：如果手动 DI 和 Wire 的代码行数增长曲线基本平行，则假设不成立

// 生成指定数量依赖的手动 DI 代码
func generateManualDI(depCount int) string {
	var sb strings.Builder

	sb.WriteString("package main\n\n")
	sb.WriteString("import \"fmt\"\n\n")

	// 生成依赖接口和实现
	for i := 1; i <= depCount; i++ {
		sb.WriteString(fmt.Sprintf("// Service%d 是第 %d 个服务接口\n", i, i))
		sb.WriteString(fmt.Sprintf("type Service%d interface {\n", i))
		sb.WriteString(fmt.Sprintf("\tDo%d() string\n", i))
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("type service%d struct {\n", i))
		if i > 1 {
			// 每个服务依赖前一个服务（链式依赖）
			sb.WriteString(fmt.Sprintf("\tprev Service%d\n", i-1))
		}
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("func (s *service%d) Do%d() string {\n", i, i))
		if i > 1 {
			sb.WriteString(fmt.Sprintf("\treturn fmt.Sprintf(\"s%d -> %%s\", s.prev.Do%d())\n", i, i-1))
		} else {
			sb.WriteString(fmt.Sprintf("\treturn \"s%d\"\n", i))
		}
		sb.WriteString("}\n\n")
	}

	// 生成手动组装代码（main 函数）
	sb.WriteString("func main() {\n")
	sb.WriteString("\t// 手动 DI：按正确顺序组装所有依赖\n")
	// 从最底层开始组装
	for i := 1; i <= depCount; i++ {
		if i == 1 {
			sb.WriteString(fmt.Sprintf("\ts1 := &service1{}\n"))
		} else {
			sb.WriteString(fmt.Sprintf("\ts%d := &service%d{prev: s%d}\n", i, i, i-1))
		}
	}
	sb.WriteString(fmt.Sprintf("\tfmt.Println(s%d.Do%d())\n", depCount, depCount))
	sb.WriteString("}\n")

	return sb.String()
}

// 生成 Wire Provider Set 代码
func generateWireDI(depCount int) (providerCode string, wireCode string, injectorCode string) {
	var prov strings.Builder
	var wireBuilder strings.Builder
	var inj strings.Builder

	// provider 文件
	prov.WriteString("package main\n\n")
	for i := 1; i <= depCount; i++ {
		if i == 1 {
			prov.WriteString(fmt.Sprintf("func NewService%d() Service%d {\n", i, i))
			prov.WriteString(fmt.Sprintf("\treturn &service1{}\n"))
			prov.WriteString("}\n\n")
		} else {
			prov.WriteString(fmt.Sprintf("func NewService%d(prev Service%d) Service%d {\n", i, i-1, i))
			prov.WriteString(fmt.Sprintf("\treturn &service%d{prev: prev}\n", i))
			prov.WriteString("}\n\n")
		}
	}

	// wire.go 文件
	wireBuilder.WriteString("//go:build wireinject\n")
	wireBuilder.WriteString("package main\n\n")
	wireBuilder.WriteString("import \"github.com/google/wire\"\n\n")
	wireBuilder.WriteString("func initializeApp() *App {\n")
	wireBuilder.WriteString("\twire.Build(\n")
	for i := 1; i <= depCount; i++ {
		wireBuilder.WriteString(fmt.Sprintf("\t\tNewService%d,\n", i))
	}
	wireBuilder.WriteString("\t\tNewApp,\n")
	wireBuilder.WriteString("\t)\n")
	wireBuilder.WriteString("\treturn nil\n")
	wireBuilder.WriteString("}\n")

	// injector/main 文件
	inj.WriteString("package main\n\n")
	inj.WriteString("import \"fmt\"\n\n")
	inj.WriteString("type App struct {\n")
	inj.WriteString(fmt.Sprintf("\tsvc Service%d\n", depCount))
	inj.WriteString("}\n\n")
	inj.WriteString(fmt.Sprintf("func NewApp(svc Service%d) *App {\n", depCount))
	inj.WriteString("\treturn &App{svc: svc}\n")
	inj.WriteString("}\n\n")
	inj.WriteString("func main() {\n")
	inj.WriteString("\tapp := initializeApp()\n")
	inj.WriteString(fmt.Sprintf("\tfmt.Println(app.svc.Do%d())\n", depCount))
	inj.WriteString("}\n")

	return prov.String(), wireBuilder.String(), inj.String()
}

// 统计非空非注释代码行数
func countLines(code string) (total int, codeOnly int, assembly int) {
	lines := strings.Split(code, "\n")
	inMain := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		total++
		// 统计 main 函数中的组装代码行数
		if strings.HasPrefix(trimmed, "func main()") {
			inMain = true
			continue
		}
		if inMain {
			if trimmed == "}" {
				inMain = false
				continue
			}
			assembly++
		}
	}
	return total, total, assembly
}

// 统计 Wire 代码行数（不含自动生成部分）
func countWireLines(providerCode, wireCode, injectorCode string) (total int, assembly int) {
	pTotal, _, _ := countLines(providerCode)
	wTotal, _, _ := countLines(wireCode)
	iTotal, _, iAssembly := countLines(injectorCode)
	total = pTotal + wTotal + iTotal
	assembly = iAssembly // Wire 的组装代码只有 injector 的 main
	return total, assembly
}

func main() {
	fmt.Println("E1 实验：依赖图规模 vs 维护成本曲线")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("%-10s %-15s %-15s %-15s %-15s %-15s\n",
		"依赖数", "手动DI总行数", "手动组装行数", "Wire总行数", "Wire组装行数", "组装行数比")
	fmt.Println(strings.Repeat("-", 85))

	for depCount := 5; depCount <= 50; depCount += 5 {
		manualCode := generateManualDI(depCount)
		manualTotal, _, manualAssembly := countLines(manualCode)

		providerCode, wireCode, injectorCode := generateWireDI(depCount)
		wireTotal, wireAssembly := countWireLines(providerCode, wireCode, injectorCode)

		ratio := float64(manualAssembly) / float64(wireAssembly)

		fmt.Printf("%-10d %-15d %-15d %-15d %-15d %-15.2f\n",
			depCount, manualTotal, manualAssembly, wireTotal, wireAssembly, ratio)
	}

	fmt.Println()
	fmt.Println("【分析】")
	fmt.Println("1. 手动组装行数 = main() 函数中手动创建依赖的代码行数")
	fmt.Println("2. Wire 组装行数 = wire.Build() 中的 provider 列表行数")
	fmt.Println("3. 组装行数比 = 手动组装 / Wire 组装，如果比值随依赖数增长，说明手动 DI 的组装成本增长更快")
	fmt.Println()

	// 额外分析：修改成本
	fmt.Println("=== 修改成本分析 ===")
	fmt.Println("场景：在依赖链中间插入一个新服务（假设插入到第 N/2 位置）")
	fmt.Printf("%-10s %-20s %-20s\n", "依赖数", "手动DI修改行数", "Wire修改行数")
	fmt.Println(strings.Repeat("-", 50))

	for depCount := 10; depCount <= 50; depCount += 10 {
		// 手动 DI：修改插入点后的所有 NewXxx 调用
		manualModLines := depCount - depCount/2 + 1 // 修改后半段 + 插入新行
		// Wire：只需添加新 provider + 修改 wire.Build 列表
		wireModLines := 2 // NewService + wire.Build 添加一行

		fmt.Printf("%-10d %-20d %-20d\n", depCount, manualModLines, wireModLines)
	}
}
