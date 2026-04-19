package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// E4 实验：三种 DI 方案启动性能对比
// 简化版：直接生成 3 个独立项目并测量

func main() {
	fmt.Println("E4 实验：三种 DI 方案启动性能对比")
	fmt.Println("========================================")
	fmt.Println()

	depCount := 20

	// 1. 手动 DI
	manualDir := "/tmp/e4-manual"
	os.RemoveAll(manualDir)
	os.MkdirAll(manualDir, 0755)
	genManual(manualDir, depCount)

	fmt.Println("--- 手动 DI ---")
	measureAndPrint(manualDir, "手动DI", depCount)

	// 2. Wire
	wireDir := "/tmp/e4-wire"
	os.RemoveAll(wireDir)
	os.MkdirAll(wireDir, 0755)
	genWire(wireDir, depCount)

	fmt.Println("--- Wire ---")
	// Wire 需要先 generate
	cmd := exec.Command(os.ExpandEnv("$HOME/go/bin/wire"))
	cmd.Dir = wireDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Wire generate 失败: %s\n%s\n", err, string(out))
		return
	}
	measureAndPrint(wireDir, "Wire", depCount)

	// 3. Fx
	fxDir := "/tmp/e4-fx"
	os.RemoveAll(fxDir)
	os.MkdirAll(fxDir, 0755)
	genFx(fxDir, depCount)

	fmt.Println("--- Fx ---")
	measureAndPrint(fxDir, "Fx", depCount)
}

func genManual(dir string, n int) {
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module e4manual\n\ngo 1.26\n"), 0644)

	var code = "package main\n\nimport \"fmt\"\n\n"
	for i := 1; i <= n; i++ {
		code += fmt.Sprintf("type S%d interface{ V() int }\n", i)
		code += fmt.Sprintf("type s%d struct", i)
		if i > 1 {
			code += fmt.Sprintf("{ p S%d }", i-1)
		} else {
			code += "{}"
		}
		code += "\n"
		code += fmt.Sprintf("func (x *s%d) V() int { return %d }\n\n", i, i)
	}
	code += "func main() {\n"
	for i := 1; i <= n; i++ {
		if i == 1 {
			code += "\tv1 := &s1{}\n"
		} else {
			code += fmt.Sprintf("\tv%d := &s%d{p: v%d}\n", i, i, i-1)
		}
	}
	code += fmt.Sprintf("\tfmt.Println(v%d.V())\n", n)
	code += "}\n"

	os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)
}

func genWire(dir string, n int) {
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module e4wire\n\ngo 1.26\n\nrequire github.com/google/wire v0.7.0\n"), 0644)

	var prov = "package main\n\nimport \"fmt\"\n\n"
	for i := 1; i <= n; i++ {
		prov += fmt.Sprintf("type S%d interface{ V() int }\n", i)
		prov += fmt.Sprintf("type s%d struct", i)
		if i > 1 {
			prov += fmt.Sprintf("{ P S%d }", i-1)
		} else {
			prov += "{}"
		}
		prov += "\n"
		if i == 1 {
			prov += fmt.Sprintf("func (x *s%d) V() int { return %d }\n", i, i)
			prov += fmt.Sprintf("func NewS%d() S%d { return &s1{} }\n\n", i, i)
		} else {
			prov += fmt.Sprintf("func (x *s%d) V() int { return %d }\n", i, i)
			prov += fmt.Sprintf("func NewS%d(p S%d) S%d { return &s%d{P: p} }\n\n", i, i-1, i, i)
		}
	}
	prov += fmt.Sprintf("type App struct{ Final S%d }\nfunc NewApp(final S%d) *App { return &App{Final: final} }\nfunc (a *App) Run() { fmt.Println(\"ok\") }\n", n, n)
	os.WriteFile(filepath.Join(dir, "provider.go"), []byte(prov), 0644)

	wireGo := "//go:build wireinject\n\npackage main\n\nimport \"github.com/google/wire\"\n\nfunc InitializeApp() *App {\n\twire.Build(\n"
	for i := 1; i <= n; i++ {
		wireGo += fmt.Sprintf("\t\tNewS%d,\n", i)
	}
	wireGo += "\t\tNewApp,\n\t)\n\treturn nil\n}\n"
	os.WriteFile(filepath.Join(dir, "wire.go"), []byte(wireGo), 0644)

	mainGo := "package main\n\nfunc main() {\n\tInitializeApp()\n}\n"
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Run()
}

func genFx(dir string, n int) {
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module e4fx\n\ngo 1.26\n\nrequire go.uber.org/fx v1.23.0\n"), 0644)

	var code = "package main\n\nimport (\n\t\"fmt\"\n\t\"go.uber.org/fx\"\n)\n\n"
	for i := 1; i <= n; i++ {
		code += fmt.Sprintf("type S%d interface{ V() int }\n", i)
		code += fmt.Sprintf("type s%d struct", i)
		if i > 1 {
			code += fmt.Sprintf("{ P S%d }", i-1)
		} else {
			code += "{}"
		}
		code += "\n"
		if i == 1 {
			code += fmt.Sprintf("func (x *s%d) V() int { return %d }\n", i, i)
			code += fmt.Sprintf("func NewS%d() S%d { return &s1{} }\n\n", i, i)
		} else {
			code += fmt.Sprintf("func (x *s%d) V() int { return %d }\n", i, i)
			code += fmt.Sprintf("func NewS%d(p S%d) S%d { return &s%d{P: p} }\n\n", i, i-1, i, i)
		}
	}
	code += "func main() {\n\tfx.New(\n\t\tfx.Provide(\n"
	for i := 1; i <= n; i++ {
		code += fmt.Sprintf("\t\t\tNewS%d,\n", i)
		_ = i
	}
	code += "\t\t),\n\t\tfx.Invoke(func() { fmt.Println(\"ok\") }),\n\t)\n}\n"

	os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)

	cmd2 := exec.Command("go", "mod", "tidy")
	cmd2.Dir = dir
	cmd2.Run()
}

func measureAndPrint(dir, name string, deps int) {
	binPath := filepath.Join(dir, "app")

	// 编译（3次取平均）
	var totalCompile time.Duration
	for i := 0; i < 3; i++ {
		start := time.Now()
		cmd := exec.Command("go", "build", "-o", binPath, ".")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("编译失败: %s\n%s\n", err, string(out))
			return
		}
		totalCompile += time.Since(start)
	}
	avgCompile := totalCompile / 3

	// 二进制大小
	info, _ := os.Stat(binPath)
	binSize := info.Size()

	// 运行（3次取平均）
	var totalRun time.Duration
	for i := 0; i < 3; i++ {
		start := time.Now()
		cmd := exec.Command(binPath)
		cmd.CombinedOutput()
		totalRun += time.Since(start)
	}
	avgRun := totalRun / 3

	fmt.Printf("  依赖数: %d | 编译: %v | 二进制: %d KB | 启动: %v\n",
		deps, avgCompile.Round(time.Millisecond), binSize/1024, avgRun.Round(time.Microsecond))
}
