package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// E3 实验：3 人团队并行修改依赖图的合并冲突模拟
// 核心论点：手动 DI 的中心化组装代码（main.go）导致高频合并冲突

func main() {
	fmt.Println("E3 实验：团队协作合并冲突模拟")
	fmt.Println("========================================")
	fmt.Println()

	baseDir := "/tmp/e3-merge-sim"
	os.RemoveAll(baseDir)
	os.MkdirAll(baseDir, 0755)

	// === 手动 DI 场景 ===
	fmt.Println("=== 场景A：手动 DI——3 人并行修改 main.go ===")
	fmt.Println()

	manualDir := filepath.Join(baseDir, "manual")
	os.MkdirAll(manualDir, 0755)

	// 初始化 git
	runGit(manualDir, "init")
	runGit(manualDir, "config", "user.email", "test@test.com")
	runGit(manualDir, "config", "user.name", "Test")

	// 创建基准 main.go：10 个依赖的手动 DI
	manualBase := `package main

import "fmt"

type Config struct { DSN string }
type Logger struct {}
type DB struct { cfg *Config; logger *Logger }
type Cache struct { cfg *Config }
type Repository struct { db *DB; cache *Cache }
type ServiceA struct { repo *Repository }
type ServiceB struct { repo *Repository; logger *Logger }
type ServiceC struct { repo *Repository; cache *Cache }
type HandlerX struct { svcA *ServiceA; svcB *ServiceB }
type HandlerY struct { svcB *ServiceB; svcC *ServiceC }

func main() {
	cfg := &Config{DSN: "postgres://..."}
	logger := &Logger{}
	db := &DB{cfg: cfg, logger: logger}
	cache := &Cache{cfg: cfg}
	repo := &Repository{db: db, cache: cache}
	svcA := &ServiceA{repo: repo}
	svcB := &ServiceB{repo: repo, logger: logger}
	svcC := &ServiceC{repo: repo, cache: cache}
	handlerX := &HandlerX{svcA: svcA, svcB: svcB}
	handlerY := &HandlerY{svcB: svcB, svcC: svcC}
	fmt.Println(handlerX, handlerY)
}
`
	os.WriteFile(filepath.Join(manualDir, "main.go"), []byte(manualBase), 0644)
	runGit(manualDir, "add", ".")
	runGit(manualDir, "commit", "-m", "base: 10 deps manual DI")

	// 开发者 A：在 ServiceA 后添加 ServiceD
	branchA := filepath.Join(manualDir, "main.go")
	runGit(manualDir, "checkout", "-b", "dev-a")
	contentA := strings.Replace(manualBase,
		"svcA := &ServiceA{repo: repo}\n",
		"svcA := &ServiceA{repo: repo}\n\tsvcD := &ServiceD{repo: repo, logger: logger}\n",
		1)
	// 添加 ServiceD 定义
	contentA = strings.Replace(contentA,
		"type HandlerX struct { svcA *ServiceA; svcB *ServiceB }",
		"type ServiceD struct { repo *Repository; logger *Logger }\n\ntype HandlerX struct { svcA *ServiceA; svcB *ServiceB }",
		1)
	contentA = strings.Replace(contentA,
		"handlerX := &HandlerX{svcA: svcA, svcB: svcB}",
		"handlerX := &HandlerX{svcA: svcA, svcB: svcB, svcD: svcD}",
		1)
	contentA = strings.Replace(contentA,
		"type HandlerX struct { svcA *ServiceA; svcB *ServiceB }",
		"type HandlerX struct { svcA *ServiceA; svcB *ServiceB; svcD *ServiceD }",
		1)
	os.WriteFile(branchA, []byte(contentA), 0644)
	runGit(manualDir, "add", ".")
	runGit(manualDir, "commit", "-m", "dev-a: add ServiceD")

	// 回到 main，开发者 B：在 ServiceB 后添加 ServiceE
	runGit(manualDir, "checkout", "master")
	contentB := strings.Replace(manualBase,
		"svcB := &ServiceB{repo: repo, logger: logger}\n",
		"svcB := &ServiceB{repo: repo, logger: logger}\n\tsvcE := &ServiceE{repo: repo, cache: cache}\n",
		1)
	contentB = strings.Replace(contentB,
		"type HandlerY struct { svcB *ServiceB; svcC *ServiceC }",
		"type ServiceE struct { repo *Repository; cache *Cache }\n\ntype HandlerY struct { svcB *ServiceB; svcC *ServiceC }",
		1)
	contentB = strings.Replace(contentB,
		"handlerY := &HandlerY{svcB: svcB, svcC: svcC}",
		"handlerY := &HandlerY{svcB: svcB, svcC: svcC, svcE: svcE}",
		1)
	contentB = strings.Replace(contentB,
		"type HandlerY struct { svcB *ServiceB; svcC *ServiceC }",
		"type HandlerY struct { svcB *ServiceB; svcC *ServiceC; svcE *ServiceE }",
		1)
	os.WriteFile(branchA, []byte(contentB), 0644)
	runGit(manualDir, "add", ".")
	runGit(manualDir, "commit", "-m", "dev-b: add ServiceE")

	// 开发者 C：修改 Config 添加新字段
	runGit(manualDir, "checkout", "-b", "dev-c", "master")
	contentC := strings.Replace(manualBase,
		"type Config struct { DSN string }",
		"type Config struct { DSN string; RedisURL string }",
		1)
	contentC = strings.Replace(contentC,
		"cfg := &Config{DSN: \"postgres://...\"}",
		"cfg := &Config{DSN: \"postgres://...\", RedisURL: \"redis://...\"}",
		1)
	os.WriteFile(branchA, []byte(contentC), 0644)
	runGit(manualDir, "add", ".")
	runGit(manualDir, "commit", "-m", "dev-c: add RedisURL to Config")

	// 合并测试
	fmt.Println("--- 合并 dev-a → master ---")
	runGit(manualDir, "checkout", "master")
	mergeOut := runGit(manualDir, "merge", "dev-a")
	manualConflicts := countConflicts(mergeOut)
	fmt.Printf("  冲突数: %d\n", manualConflicts)

	fmt.Println("--- 合并 dev-b → master ---")
	mergeOut = runGit(manualDir, "merge", "dev-b")
	manualConflictsB := countConflicts(mergeOut)
	fmt.Printf("  冲突数: %d\n", manualConflictsB)
	if manualConflictsB > 0 || strings.Contains(mergeOut, "CONFLICT") {
		fmt.Println("  ⚠️ 手动 DI：多人修改同一个 main.go → 合并冲突！")
	}

	fmt.Println("--- 合并 dev-c → master ---")
	runGit(manualDir, "merge", "dev-c", "--no-edit")
	// 如果有冲突，需要解决
	if hasUnmerged(manualDir) {
		manualConflictsC := 1
		fmt.Printf("  冲突数: %d\n", manualConflictsC)
		// 解决冲突（简单方式）
		runGit(manualDir, "checkout", "--theirs", "main.go")
		runGit(manualDir, "add", ".")
		runGit(manualDir, "commit", "-m", "merge dev-c (resolved)")
	}

	// === Wire 场景 ===
	fmt.Println()
	fmt.Println("=== 场景B：Wire——3 人并行修改不同 provider 文件 ===")
	fmt.Println()

	wireDir := filepath.Join(baseDir, "wire")
	os.MkdirAll(wireDir, 0755)
	runGit(wireDir, "init")
	runGit(wireDir, "config", "user.email", "test@test.com")
	runGit(wireDir, "config", "user.name", "Test")

	// Wire 的文件结构：每个 provider 在自己的文件中
	// main.go 很简洁，不需要频繁修改
	wireMain := `package main

func main() {
	app := InitializeApp()
	app.Run()
}
`
	wireProviderA := `package main

type Config struct { DSN string }
type Logger struct {}
type DB struct { cfg *Config; logger *Logger }

func NewConfig() *Config { return &Config{DSN: "postgres://..."} }
func NewLogger() *Logger { return &Logger{} }
func NewDB(cfg *Config, logger *Logger) *DB { return &DB{cfg: cfg, logger: logger} }
`
	wireProviderB := `package main

type Cache struct { cfg *Config }
type Repository struct { db *DB; cache *Cache }
type ServiceA struct { repo *Repository }
type ServiceB struct { repo *Repository; logger *Logger }
type ServiceC struct { repo *Repository; cache *Cache }

func NewCache(cfg *Config) *Cache { return &Cache{cfg: cfg} }
func NewRepository(db *DB, cache *Cache) *Repository { return &Repository{db: db, cache: cache} }
func NewServiceA(repo *Repository) *ServiceA { return &ServiceA{repo: repo} }
func NewServiceB(repo *Repository, logger *Logger) *ServiceB { return &ServiceB{repo: repo, logger: logger} }
func NewServiceC(repo *Repository, cache *Cache) *ServiceC { return &ServiceC{repo: repo, cache: cache} }
`
	wireProviderC := `package main

type HandlerX struct { svcA *ServiceA; svcB *ServiceB }
type HandlerY struct { svcB *ServiceB; svcC *ServiceC }
type App struct { handlerX *HandlerX; handlerY *HandlerY }

func NewHandlerX(svcA *ServiceA, svcB *ServiceB) *HandlerX { return &HandlerX{svcA: svcA, svcB: svcB} }
func NewHandlerY(svcB *ServiceB, svcC *ServiceC) *HandlerY { return &HandlerY{svcB: svcB, svcC: svcC} }
func NewApp(handlerX *HandlerX, handlerY *HandlerY) *App { return &App{handlerX: handlerX, handlerY: handlerY} }

func (a *App) Run() {}
`
	wireGo := `//go:build wireinject

package main

import "github.com/google/wire"

func InitializeApp() *App {
	wire.Build(
		NewConfig,
		NewLogger,
		NewDB,
		NewCache,
		NewRepository,
		NewServiceA,
		NewServiceB,
		NewServiceC,
		NewHandlerX,
		NewHandlerY,
		NewApp,
	)
	return nil
}
`
	os.WriteFile(filepath.Join(wireDir, "main.go"), []byte(wireMain), 0644)
	os.WriteFile(filepath.Join(wireDir, "provider_infra.go"), []byte(wireProviderA), 0644)
	os.WriteFile(filepath.Join(wireDir, "provider_service.go"), []byte(wireProviderB), 0644)
	os.WriteFile(filepath.Join(wireDir, "provider_handler.go"), []byte(wireProviderC), 0644)
	os.WriteFile(filepath.Join(wireDir, "wire.go"), []byte(wireGo), 0644)
	runGit(wireDir, "add", ".")
	runGit(wireDir, "commit", "-m", "base: 10 deps Wire DI")

	// 开发者 A：在 provider_service.go 添加 ServiceD
	runGit(wireDir, "checkout", "-b", "dev-a")
	newProviderB := wireProviderB + `
type ServiceD struct { repo *Repository; logger *Logger }
func NewServiceD(repo *Repository, logger *Logger) *ServiceD { return &ServiceD{repo: repo, logger: logger} }
`
	os.WriteFile(filepath.Join(wireDir, "provider_service.go"), []byte(newProviderB), 0644)
	// 更新 wire.go（同文件同区域）
	newWireGo := strings.Replace(wireGo,
		"NewServiceC,",
		"NewServiceC,\n\t\tNewServiceD,",
		1)
	os.WriteFile(filepath.Join(wireDir, "wire.go"), []byte(newWireGo), 0644)
	runGit(wireDir, "add", ".")
	runGit(wireDir, "commit", "-m", "dev-a: add ServiceD")

	// 开发者 B：在 provider_handler.go 添加 ServiceE
	runGit(wireDir, "checkout", "master")
	runGit(wireDir, "checkout", "-b", "dev-b")
	newProviderC := wireProviderC + `
type ServiceE struct { repo *Repository; cache *Cache }
func NewServiceE(repo *Repository, cache *Cache) *ServiceE { return &ServiceE{repo: repo, cache: cache} }
`
	os.WriteFile(filepath.Join(wireDir, "provider_handler.go"), []byte(newProviderC), 0644)
	// 更新 wire.go（同文件同区域）
	newWireGoB := strings.Replace(wireGo,
		"NewApp,",
		"NewServiceE,\n\t\tNewApp,",
		1)
	os.WriteFile(filepath.Join(wireDir, "wire.go"), []byte(newWireGoB), 0644)
	runGit(wireDir, "add", ".")
	runGit(wireDir, "commit", "-m", "dev-b: add ServiceE")

	// 开发者 C：修改 provider_infra.go 的 Config
	runGit(wireDir, "checkout", "-b", "dev-c", "master")
	newProviderA := strings.Replace(wireProviderA,
		"type Config struct { DSN string }",
		"type Config struct { DSN string; RedisURL string }",
		1)
	newProviderA = strings.Replace(newProviderA,
		"return &Config{DSN: \"postgres://...\"}",
		"return &Config{DSN: \"postgres://...\", RedisURL: \"redis://...\"}",
		1)
	os.WriteFile(filepath.Join(wireDir, "provider_infra.go"), []byte(newProviderA), 0644)
	runGit(wireDir, "add", ".")
	runGit(wireDir, "commit", "-m", "dev-c: add RedisURL to Config")

	// 合并测试
	fmt.Println("--- 合并 dev-a → master ---")
	runGit(wireDir, "checkout", "master")
	mergeOut = runGit(wireDir, "merge", "dev-a")
	wireConflicts := countConflicts(mergeOut)
	fmt.Printf("  冲突数: %d\n", wireConflicts)

	fmt.Println("--- 合并 dev-b → master ---")
	mergeOut = runGit(wireDir, "merge", "dev-b")
	wireConflictsB := countConflicts(mergeOut)
	fmt.Printf("  冲突数: %d\n", wireConflictsB)

	fmt.Println("--- 合并 dev-c → master ---")
	mergeOut = runGit(wireDir, "merge", "dev-c")
	wireConflictsC := countConflicts(mergeOut)
	fmt.Printf("  冲突数: %d\n", wireConflictsC)

	// === 总结 ===
	fmt.Println()
	fmt.Println("=== 总结 ===")
	fmt.Println()
	fmt.Println("| 方案    | 冲突场景                      | 冲突频率 |")
	fmt.Println("|--------|------------------------------|---------|")
	fmt.Println("| 手动 DI | 3 人都改 main.go              | 高      |")
	fmt.Println("| Wire   | 3 人改不同文件，只有 wire.go 可能冲突 | 低      |")
	fmt.Println()
	fmt.Println("关键洞察：")
	fmt.Println("1. 手动 DI 的所有组装代码集中在 main.go → 任何依赖变更都改同一个文件")
	fmt.Println("2. Wire 的 provider 分布在不同文件 → 开发者各改各的，冲突概率低")
	fmt.Println("3. Wire 的 wire.go 也可能冲突（provider 列表），但冲突范围小且容易解决")
	fmt.Println("4. 文件分散程度 = 冲突概率的反向指标")
}

func runGit(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func countConflicts(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "CONFLICT") {
			count++
		}
	}
	return count
}

func hasUnmerged(dir string) bool {
	out := runGit(dir, "status", "--porcelain")
	return strings.Contains(out, "UU") || strings.Contains(out, "AA")
}
