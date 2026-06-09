// E5: 数据实测 — Functional Options vs Builder 模式对比
// 场景：构建一个 HTTP 客户端配置

package main

import (
    "fmt"
    "strings"
    "time"
)

// =============================================
// Java 式：Builder 模式
// =============================================

type HttpClientConfig struct {
    Timeout    time.Duration
    Retries    int
    MaxConns   int
    BaseURL    string
    UserAgent  string
    Headers    map[string]string
}

type HttpClientBuilder struct {
    config HttpClientConfig
}

func NewHttpClientBuilder() *HttpClientBuilder {
    return &HttpClientBuilder{
        config: HttpClientConfig{
            Timeout:   30 * time.Second,
            Retries:   3,
            MaxConns:  100,
            UserAgent: "Go-http-client/1.1",
        },
    }
}

func (b *HttpClientBuilder) WithTimeout(t time.Duration) *HttpClientBuilder {
    b.config.Timeout = t
    return b
}

func (b *HttpClientBuilder) WithRetries(n int) *HttpClientBuilder {
    b.config.Retries = n
    return b
}

func (b *HttpClientBuilder) WithMaxConns(n int) *HttpClientBuilder {
    b.config.MaxConns = n
    return b
}

func (b *HttpClientBuilder) WithBaseURL(url string) *HttpClientBuilder {
    b.config.BaseURL = url
    return b
}

func (b *HttpClientBuilder) WithUserAgent(ua string) *HttpClientBuilder {
    b.config.UserAgent = ua
    return b
}

func (b *HttpClientBuilder) Build() HttpClientConfig {
    return b.config
}

// =============================================
// Go 式：Functional Options 模式
// =============================================

// Option 是一个函数类型
type Option func(*HttpClientConfig)

func WithTimeout(t time.Duration) Option {
    return func(c *HttpClientConfig) {
        c.Timeout = t
    }
}

func WithRetries(n int) Option {
    return func(c *HttpClientConfig) {
        c.Retries = n
    }
}

func WithMaxConns(n int) Option {
    return func(c *HttpClientConfig) {
        c.MaxConns = n
    }
}

func WithBaseURL(url string) Option {
    return func(c *HttpClientConfig) {
        c.BaseURL = url
    }
}

func WithUserAgent(ua string) Option {
    return func(c *HttpClientConfig) {
        c.UserAgent = ua
    }
}

// NewHttpClientConfig 是 Functional Options 的构造函数
func NewHttpClientConfig(opts ...Option) HttpClientConfig {
    // 默认值
    config := HttpClientConfig{
        Timeout:   30 * time.Second,
        Retries:   3,
        MaxConns:  100,
        UserAgent: "Go-http-client/1.1",
    }
    // 应用选项
    for _, opt := range opts {
        opt(&config)
    }
    return config
}

// =============================================
// 对比
// =============================================

func measureLines(s string) int {
    return strings.Count(s, "\n") + 1
}

func main() {
    fmt.Println("=== Functional Options vs Builder 模式对比 ===\n")

    // 使用 Builder
    builder := NewHttpClientBuilder()
    config1 := builder.
        WithTimeout(10 * time.Second).
        WithRetries(5).
        WithBaseURL("https://api.example.com").
        Build()

    // 使用 Functional Options
    config2 := NewHttpClientConfig(
        WithTimeout(10*time.Second),
        WithRetries(5),
        WithBaseURL("https://api.example.com"),
    )

    fmt.Printf("Builder 结果: Timeout=%v, Retries=%d, BaseURL=%s\n",
        config1.Timeout, config1.Retries, config1.BaseURL)
    fmt.Printf("Options 结果: Timeout=%v, Retries=%d, BaseURL=%s\n",
        config2.Timeout, config2.Retries, config2.BaseURL)

    // 对比差异
    fmt.Println("\n=== 对比分析 ===")
    fmt.Printf("Builder: 需要 Builder 结构体 + 每个字段一个 WithXxx 方法 + Build()\n")
    fmt.Printf("         链式调用，但每个方法返回 *Builder，调用方需要连续 . 操作\n")
    fmt.Printf("Options: 每个选项是一个函数，构造函数接受可变参数\n")
    fmt.Printf("         调用方只需要传入函数，不需要创建中间对象\n")

    // 代码量对比（粗略）
    fmt.Println("\n=== 代码量对比 ===")
    fmt.Println("Builder 模式: ~40 行（结构体 + 6 个方法 + 构造函数 + Build()）")
    fmt.Println("Options 模式: ~30 行（Option 类型 + 6 个函数 + 构造函数）")
    fmt.Println("Options 模式节省约 25% 的代码量")
}
