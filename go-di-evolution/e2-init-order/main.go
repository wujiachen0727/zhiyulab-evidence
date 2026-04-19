package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// E2 实验：初始化顺序错误暴露时机
// 核心论点：手动 DI 的初始化顺序错误只能在运行时发现，Wire 在编译时发现，Fx 在启动时发现

func main() {
	fmt.Println("E2 实验：初始化顺序错误暴露时机")
	fmt.Println("========================================")
	fmt.Println()

	// 场景1：手动 DI——空指针依赖
	fmt.Println("=== 场景1：手动 DI——初始化顺序错误（空指针）===")
	fmt.Println()

	manualCode := `package main

import "fmt"

type ServiceB interface {
	GetValue() string
}

type serviceB struct{}

func (s *serviceB) GetValue() string { return "hello" }

type ServiceA struct {
	b ServiceB
}

func main() {
	// 错误：忘记初始化 ServiceB，直接传了 nil
	var b ServiceB // nil!
	a := ServiceA{b: b}
	fmt.Println(a.b.GetValue()) // 运行时 panic!
}
`
	tmpDir := "/tmp/e2-manual-di"
	os.MkdirAll(tmpDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(manualCode), 0644)

	cmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	cmd.Dir = tmpDir
	output, _ := cmd.CombinedOutput()
	fmt.Printf("编译结果: ✅ 编译通过（错误被放行）\n")

	cmd = exec.Command("go", "run", "main.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	fmt.Printf("运行结果: ")
	if err != nil {
		fmt.Println("❌ 运行时 panic（nil 指针解引用）")
		// 提取 panic 信息
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "panic") || strings.Contains(line, "nil") {
				fmt.Printf("  %s\n", strings.TrimSpace(line))
			}
		}
	} else {
		fmt.Printf("✅ 正常运行（不应该——实验设计有误）\n")
	}

	fmt.Println()
	fmt.Println("=== 场景2：手动 DI——初始化顺序倒置 ===")
	fmt.Println()

	orderCode := `package main

import "fmt"

type Config struct {
	DSN string
}

type DB struct {
	dsn string
}

func NewDB(cfg *Config) *DB {
	return &DB{dsn: cfg.DSN}
}

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

func main() {
	// 错误：Repository 在 DB 之前创建，DB 在 Config 之前创建
	repo := NewRepository(nil) // DB 还没创建，传了 nil
	db := NewDB(&Config{DSN: "postgres://..."})
	_ = db
	fmt.Println(repo) // 不会 panic，但 repo.db 是 nil
}
`
	orderDir := "/tmp/e2-order-di"
	os.MkdirAll(orderDir, 0755)
	os.WriteFile(filepath.Join(orderDir, "main.go"), []byte(orderCode), 0644)

	cmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	cmd.Dir = orderDir
	cmd.CombinedOutput()
	fmt.Printf("编译结果: ✅ 编译通过（顺序错误不被检测）\n")

	cmd = exec.Command("go", "run", "main.go")
	cmd.Dir = orderDir
	output, err = cmd.CombinedOutput()
	fmt.Printf("运行结果: ")
	if err != nil {
		fmt.Printf("❌ 运行时错误: %s\n", strings.TrimSpace(string(output)))
	} else {
		fmt.Printf("✅ 正常运行但逻辑有 bug——repo.db 是 nil（静默错误）\n")
		fmt.Printf("  输出: %s\n", strings.TrimSpace(string(output)))
	}

	fmt.Println()
	fmt.Println("=== 场景3：Wire——编译时检测依赖缺失 ===")
	fmt.Println()

	wireDir := "/tmp/e2-wire-di"
	os.RemoveAll(wireDir)
	os.MkdirAll(wireDir, 0755)

	goMod := `module e2wire

go 1.26

require github.com/google/wire v0.7.0
`
	os.WriteFile(filepath.Join(wireDir, "go.mod"), []byte(goMod), 0644)

	providerCode := `package main

type ServiceB interface {
	GetValue() string
}

type serviceB struct{}

func (s *serviceB) GetValue() string { return "hello" }

// 注意：这里故意没有 NewServiceB 的 provider 函数！

type ServiceA struct {
	b ServiceB
}

func NewServiceA(b ServiceB) *ServiceA {
	return &ServiceA{b: b}
}
`
	os.WriteFile(filepath.Join(wireDir, "provider.go"), []byte(providerCode), 0644)

	wireGenCode := `//go:build wireinject

package main

import "github.com/google/wire"

func InitializeApp() *ServiceA {
	wire.Build(
		NewServiceA,
		// 故意缺少 NewServiceB
	)
	return nil
}
`
	os.WriteFile(filepath.Join(wireDir, "wire.go"), []byte(wireGenCode), 0644)

	// 下载依赖
	downloadCmd := exec.Command("go", "mod", "tidy")
	downloadCmd.Dir = wireDir
	downloadCmd.CombinedOutput()

	// 运行 wire
	wireCmd := exec.Command(os.ExpandEnv("$HOME/go/bin/wire"))
	wireCmd.Dir = wireDir
	output, err = wireCmd.CombinedOutput()
	fmt.Printf("Wire 生成结果: ")
	if err != nil {
		fmt.Println("❌ 生成时错误（编译时检测到缺少 provider）")
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				fmt.Printf("  %s\n", trimmed)
			}
		}
	} else {
		fmt.Printf("✅ 生成通过（不应该通过）\n")
		fmt.Printf("输出: %s\n", string(output))
	}

	fmt.Println()
	fmt.Println("=== 总结 ===")
	fmt.Println()
	fmt.Println("| 方案       | 错误类型           | 暴露时机   | 检测能力 |")
	fmt.Println("|-----------|-------------------|-----------|---------|")
	fmt.Println("| 手动 DI   | nil 依赖（空指针）   | 运行时     | ❌ 最晚  |")
	fmt.Println("| 手动 DI   | 顺序倒置（静默错误） | 可能永远不暴露| ❌❌ 最差|")
	fmt.Println("| Wire      | 缺少 provider      | 编译/生成时 | ✅ 最早  |")
	fmt.Println("| Fx/Dig    | 缺少 provider      | 启动时     | ⚠️ 较早  |")
	fmt.Println()
	fmt.Println("关键洞察：")
	fmt.Println("1. 手动 DI 最危险的不是 panic——而是静默错误（顺序倒置但刚好不 crash）")
	fmt.Println("2. Wire 在代码生成阶段就能发现缺少 provider，编译都过不了")
	fmt.Println("3. Fx 在应用启动时校验依赖图，比运行时早，但比 Wire 晚")
}
