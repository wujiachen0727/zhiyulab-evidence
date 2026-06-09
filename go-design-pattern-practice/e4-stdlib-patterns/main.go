// E4: Go 标准库中的设计模式实例
// 不是"标准库实现了模式"，而是"标准库的设计意图正好匹配了模式"

package main

import (
    "bytes"
    "context"
    "crypto/md5"
    "fmt"
    "io"
    "net/http"
    "sync"
)

// =============================================
// 1. 单例模式：sync.Once
// 标准库没有"全局唯一实例"的强制约束
// 但通过 sync.Once 提供了"只执行一次"的语义
// =============================================

func demonstrateOnce() {
    var once sync.Once
    var config map[string]string

    loadConfig := func() {
        config = map[string]string{"key": "value"}
        fmt.Println("  [sync.Once] 配置已加载（只会执行一次）")
    }

    // 多次调用，loadConfig 只执行一次
    for i := 0; i < 3; i++ {
        once.Do(loadConfig)
    }
    fmt.Printf("  config: %v\n", config)
}

// =============================================
// 2. 工厂模式：http.NewRequest, md5.New, bytes.NewReader
// Go 的"工厂"通常是 NewXxx 函数
// 区别：不是"根据条件创建不同对象"，而是"封装初始化逻辑"
// =============================================

func demonstrateFactory() {
    // 标准库中的"简单工厂"
    req, _ := http.NewRequest("GET", "https://example.com", nil)
    hash := md5.New()
    reader := bytes.NewReader([]byte("hello"))

    fmt.Printf("  [工厂模式] req.Method=%s, hash.Type()=%T, reader.Size()=%d\n",
        req.Method, hash, reader.Size())
}

// =============================================
// 3. 策略模式：http.Handler
// Handler 接口定义了"处理请求"的策略
// 不同的 Handler 实现就是不同的策略
// =============================================

// HandlerFunc 是一个策略实现
// 组合多个 Handler（中间件）是策略模式的另一种形式

func demonstrateStrategy() {
    // Handler 接口 = 策略接口
    // HandlerFunc = 一个具体策略
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("hello"))
    })
    fmt.Printf("  [策略模式] handler=%T\n", handler)
}

// =============================================
// 4. 适配器模式：http.HandlerFunc
// 将普通函数适配为 Handler 接口
// =============================================

func demonstrateAdapter() {
    // HandlerFunc 是一个适配器：
    // 将 func(ResponseWriter, *Request) 适配为 Handler 接口
    myHandler := func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("adapted"))
    }
    adapted := http.HandlerFunc(myHandler)
    fmt.Printf("  [适配器模式] adapted=%T\n", adapted)
}

// =============================================
// 5. 装饰器模式：io.MultiReader, io.TeeReader
// 包装一个 io.Reader 增加功能
// =============================================

func demonstrateDecorator() {
    r1 := bytes.NewReader([]byte("part1 "))
    r2 := bytes.NewReader([]byte("part2"))
    multi := io.MultiReader(r1, r2)

    data, _ := io.ReadAll(multi)
    fmt.Printf("  [装饰器模式] MultiReader: %s\n", string(data))
}

// =============================================
// 6. 观察者模式：context.Context 的取消传播
// 父 context 取消时，所有子 context 收到通知
// =============================================

func demonstrateObserver() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 子 context 观察父 context 的状态
    childCtx, childCancel := context.WithCancel(ctx)
    defer childCancel()

    // 父 context 取消时，子 context 也会收到取消信号
    cancel()

    select {
    case <-childCtx.Done():
        fmt.Printf("  [观察者模式] 子 context 收到取消通知: %v\n", childCtx.Err())
    default:
        fmt.Println("  子 context 未取消")
    }
}

// =============================================
// 7. 建造者模式：strings.Builder, bytes.Buffer
// 分步构建一个复杂对象
// =============================================

func demonstrateBuilder() {
    var buf bytes.Buffer
    buf.WriteString("Hello")
    buf.WriteByte(' ')
    buf.WriteString("World")
    fmt.Printf("  [建造者模式] Builder result: %q\n", buf.String())
}

func main() {
    fmt.Println("=== Go 标准库中的设计模式 ===")
    fmt.Println()

    fmt.Println("1. 单例模式 (sync.Once):")
    demonstrateOnce()
    fmt.Println()

    fmt.Println("2. 工厂模式 (NewXxx 函数):")
    demonstrateFactory()
    fmt.Println()

    fmt.Println("3. 策略模式 (http.Handler):")
    demonstrateStrategy()
    fmt.Println()

    fmt.Println("4. 适配器模式 (http.HandlerFunc):")
    demonstrateAdapter()
    fmt.Println()

    fmt.Println("5. 装饰器模式 (io.MultiReader):")
    demonstrateDecorator()
    fmt.Println()

    fmt.Println("6. 观察者模式 (context.Context):")
    demonstrateObserver()
    fmt.Println()

    fmt.Println("7. 建造者模式 (bytes.Buffer):")
    demonstrateBuilder()
    fmt.Println()

    fmt.Println("=== 关键洞察 ===")
    fmt.Println("Go 标准库没有'实现设计模式'这个目标")
    fmt.Println("而是每个设计模式对应的设计意图，在 Go 的语境下找到了更简洁的表达")
    fmt.Println("比如：观察者模式 → context 取消传播")
    fmt.Println("      策略模式 → http.Handler 接口")
    fmt.Println("      适配器模式 → http.HandlerFunc 函数适配")
}
