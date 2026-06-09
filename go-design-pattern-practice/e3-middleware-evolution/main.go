// E3: 场景模拟 — HTTP 中间件链的自然演化
// 展示"策略模式 + 装饰器模式"在 Go 中的自然涌现
// 场景：从一个简单的 HTTP 服务器，逐步添加中间件

package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
)

// =============================================
// 阶段 1：简单的 HTTP 处理器
// =============================================

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func phase1() {
    fmt.Println("=== 阶段 1：简单处理器 ===")
    http.HandleFunc("/hello", helloHandler)
    fmt.Println("  服务器已启动（注释掉以防止阻塞）")
}

// =============================================
// 阶段 2：需求变化 — 需要记录日志
// 最直接的方式：在每个 Handler 里加日志
// 问题：代码重复、横切关注点
// =============================================

func helloHandlerWithLog(w http.ResponseWriter, r *http.Request) {
    log.Printf("[%s] %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
    fmt.Fprintf(w, "Hello, World!")
}

// =============================================
// 阶段 3：Go 式解法 — 中间件模式
// 装饰器模式自然涌现
// =============================================

// Middleware 是一个函数类型，接受 Handler 返回 Handler
// 这是装饰器模式在 Go 中的标准形式
type Middleware func(http.Handler) http.Handler

// LoggerMiddleware 是一个具体的装饰器
func LoggerMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("[请求开始] %s %s", r.Method, r.URL.Path)

        next.ServeHTTP(w, r) // 调用被装饰的 Handler

        log.Printf("[请求结束] %s %s 耗时: %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// AuthMiddleware 另一个装饰器
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        // 验证 token...
        next.ServeHTTP(w, r)
    })
}

// RateLimitMiddleware 又一个装饰器
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 模拟限流检查
        // 实际项目中会用 sliding window / token bucket
        next.ServeHTTP(w, r)
    })
}

// ApplyMiddleware 将多个中间件组合成一个
// 这是策略模式：选择不同的中间件组合 = 选择不同的处理策略
func ApplyMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

func phase3() {
    fmt.Println("\n=== 阶段 3：装饰器 + 策略模式 ===")

    // 创建处理器
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    // 选择中间件组合 = 选择策略
    // 开发环境：只需要日志
    devHandler := ApplyMiddleware(handler, LoggerMiddleware)

    // 生产环境：日志 + 认证 + 限流
    prodHandler := ApplyMiddleware(handler, LoggerMiddleware, AuthMiddleware, RateLimitMiddleware)

    fmt.Printf("  开发环境中间件数: 1 (Logger)\n")
    fmt.Printf("  生产环境中间件数: 3 (Logger + Auth + RateLimit)\n")
    fmt.Printf("  Go 式设计：Handler 只关注业务逻辑，横切关注点由中间件处理\n")
    fmt.Printf("  这本质上就是 装饰器模式 — 包装 Handler 增加功能\n")
    fmt.Printf("  同时这也是 策略模式 — 不同的中间件组合 = 不同的处理策略\n")

    _ = devHandler
    _ = prodHandler
}

func main() {
    phase1()
    phase3()

    fmt.Println("\n=== 关键洞察 ===")
    fmt.Println("Go 的 http.Handler 接口只有一个方法: ServeHTTP")
    fmt.Println("这个极小的接口使得装饰器模式（中间件）几乎不需要任何样板代码")
    fmt.Println("在 Java 中实现同样的效果需要：Filter 接口 + FilterChain + 配置")
    fmt.Println("在 Go 中：一个函数签名就够了")
}
