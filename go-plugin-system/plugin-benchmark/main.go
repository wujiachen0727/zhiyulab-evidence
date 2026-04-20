package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Go 插件方案 5 方案性能 Benchmark
// 测试方法：每种方案执行相同的"加法运算"，测量每次调用延迟
// 环境：Go 1.26.2, darwin/arm64, Apple M3 Max
// 运行方式：go run main.go

func main() {
	fmt.Println("=== Go 插件方案性能 Benchmark ===")
	fmt.Printf("测试时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	iterations := 100000
	warmup := 1000

	type benchResult struct {
		name string
		nsOp float64
	}
	var results []benchResult

	// 方案 1：原生函数调用（基准线）
	ns := benchNative(iterations, warmup)
	results = append(results, benchResult{"1-原生调用", ns})

	// 方案 2：接口调用（模拟 plugin.Lookup 后调用路径）
	ns = benchInterfaceCall(iterations, warmup)
	results = append(results, benchResult{"2-plugin接口调用", ns})

	// 方案 3：yaegi 解释器调用
	ns = benchYaegi(iterations, warmup)
	results = append(results, benchResult{"3-yaegi解释器", ns})

	// 方案 4：wazero WASM 调用
	ns = benchWazero(iterations, warmup)
	results = append(results, benchResult{"4-wazero WASM", ns})

	// 方案 5：管道 IPC 模拟（模拟 go-plugin RPC 开销）
	ns = benchPipeIPC(iterations, warmup)
	results = append(results, benchResult{"5-管道IPC模拟", ns})

	// 汇总
	fmt.Println("\n=== 汇总 ===")
	fmt.Printf("| 方案 | 延迟 (ns/op) | 相对原生 |\n")
	fmt.Printf("|------|-------------:|--------:|\n")
	baseline := results[0].nsOp
	for _, r := range results {
		ratio := r.nsOp / baseline
		fmt.Printf("| %s | %.1f | %.1fx |\n", r.name, r.nsOp, ratio)
	}
}

// ==================== 方案 1：原生调用 ====================

func benchNative(iterations, warmup int) float64 {
	for i := 0; i < warmup; i++ {
		nativeAdd(1, 2)
	}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		nativeAdd(1, 2)
	}
	elapsed := time.Since(start)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)
	fmt.Printf("方案 1: 原生函数调用 → %.2f ns/op\n", nsPerOp)
	return nsPerOp
}

//go:noinline
func nativeAdd(a, b int) int {
	return a + b
}

// ==================== 方案 2：接口调用（模拟 plugin.Lookup） ====================

type Adder interface {
	Add(a, b int) int
}

type nativeAdderImpl struct{}

func (n *nativeAdderImpl) Add(a, b int) int { return a + b }

func benchInterfaceCall(iterations, warmup int) float64 {
	var adder Adder = &nativeAdderImpl{}
	for i := 0; i < warmup; i++ {
		adder.Add(1, 2)
	}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		adder.Add(1, 2)
	}
	elapsed := time.Since(start)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)
	fmt.Printf("方案 2: 接口调用（模拟 plugin.Lookup） → %.2f ns/op\n", nsPerOp)
	return nsPerOp
}

// ==================== 方案 3：yaegi 解释器 ====================

func benchYaegi(iterations, warmup int) float64 {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	// 定义 Add 函数
	_, err := i.Eval(`func Add(a, b int) int { return a + b }`)
	if err != nil {
		fmt.Printf("方案 3: yaegi → Eval 失败: %v\n", err)
		return -1
	}

	// 取出函数值
	v, err := i.Eval(`Add`)
	if err != nil {
		fmt.Printf("方案 3: yaegi → 取函数失败: %v\n", err)
		return -1
	}

	addFn := v.Interface().(func(int, int) int)

	// 预热
	for i := 0; i < warmup; i++ {
		addFn(1, 2)
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		addFn(1, 2)
	}
	elapsed := time.Since(start)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)
	fmt.Printf("方案 3: yaegi 解释器调用 → %.2f ns/op\n", nsPerOp)
	return nsPerOp
}

// ==================== 方案 4：wazero WASM ====================

func benchWazero(iterations, warmup int) float64 {
	// 内嵌一个最小的 WASM 模块（add 函数）
	// 用 wazero 的内置编译器方式创建
	ctx := context.Background()

	// 使用 wazero 编译内嵌 WASM
	wasmBin := createAddWasm()

	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	mod, err := r.CompileModule(ctx, wasmBin)
	if err != nil {
		fmt.Printf("方案 4: wazero → 编译失败: %v\n", err)
		return -1
	}

	inst, err := r.InstantiateModule(ctx, mod, wazero.NewModuleConfig())
	if err != nil {
		fmt.Printf("方案 4: wazero → 实例化失败: %v\n", err)
		return -1
	}
	defer inst.Close(ctx)

	addFn := inst.ExportedFunction("add")
	if addFn == nil {
		fmt.Printf("方案 4: wazero → 未找到 add 导出函数\n")
		return -1
	}

	// 预热
	for i := 0; i < warmup; i++ {
		addFn.Call(ctx, 1, 2)
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		addFn.Call(ctx, 1, 2)
	}
	elapsed := time.Since(start)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)
	fmt.Printf("方案 4: wazero WASM 调用 → %.2f ns/op\n", nsPerOp)
	return nsPerOp
}

// 创建最小的 WASM 模块（add 函数）
// WAT 格式：(module (func (export "add") (param i32 i32) (result i32) local.get 0 local.get 1 i32.add))
func createAddWasm() []byte {
	// 最小 WASM 二进制：add 函数，两个 i32 参数，返回 i32 和
	// 手工编码的 WASM 模块
	return []byte{
		0x00, 0x61, 0x73, 0x6d, // 魔数 \0asm
		0x01, 0x00, 0x00, 0x00, // 版本 1
		// 类型段
		0x01, 0x07, // section id=1, size=7
		0x01,       // 1 个类型
		0x60,       // func type
		0x02, 0x7f, 0x7f, // 2 个 i32 参数
		0x01, 0x7f, // 1 个 i32 返回值
		// 函数段
		0x03, 0x02, // section id=3, size=2
		0x01,       // 1 个函数
		0x00,       // 类型索引 0
		// 导出段
		0x07, 0x07, // section id=7, size=7
		0x01,       // 1 个导出
		0x03,       // 名称长度 3
		0x61, 0x64, 0x64, // "add"
		0x00, // func 导出
		0x00, // 函数索引 0
		// 代码段
		0x0a, 0x09, // section id=10, size=9
		0x01,       // 1 个函数体
		0x07,       // 函数体大小 7
		0x00,       // 0 个局部变量
		0x20, 0x00, // local.get 0
		0x20, 0x01, // local.get 1
		0x6a,       // i32.add
		0x0b,       // end
	}
}

// ==================== 方案 5：管道 IPC（模拟 go-plugin RPC） ====================

func benchPipeIPC(iterations, warmup int) float64 {
	// r1/w1: 主进程 → 插件进程（请求管道）
	// r2/w2: 插件进程 → 主进程（响应管道）
	r1, w1, _ := os.Pipe()
	r2, w2, _ := os.Pipe()

	// 启动 goroutine 模拟插件进程：从 r1 读取请求，计算后写入 w2
	go func() {
		reqBuf := make([]byte, 8)
		respBuf := make([]byte, 4)
		for {
			_, err := io.ReadFull(r1, reqBuf)
			if err != nil {
				return
			}
			a := int(binary.LittleEndian.Uint32(reqBuf[0:4]))
			b := int(binary.LittleEndian.Uint32(reqBuf[4:8]))
			result := uint32(a + b)
			binary.LittleEndian.PutUint32(respBuf, result)
			w2.Write(respBuf)
		}
	}()

	req := make([]byte, 8)
	binary.LittleEndian.PutUint32(req[0:4], 1)
	binary.LittleEndian.PutUint32(req[4:8], 2)
	resp := make([]byte, 4)

	// 预热
	for i := 0; i < warmup; i++ {
		w1.Write(req)
		io.ReadFull(r2, resp)
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		w1.Write(req)
		io.ReadFull(r2, resp)
	}
	elapsed := time.Since(start)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)

	// 关闭管道，让 goroutine 退出
	w1.Close()
	r1.Close()
	w2.Close()
	r2.Close()

	fmt.Printf("方案 5: 管道 IPC 模拟 → %.0f ns/op\n", nsPerOp)
	fmt.Printf("  真实 go-plugin (gRPC+Protobuf) 约 30000-50000 ns/op\n")
	return nsPerOp
}
