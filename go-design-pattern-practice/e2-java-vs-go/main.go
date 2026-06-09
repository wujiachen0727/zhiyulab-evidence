// E2: 同一功能 — Java式(继承) vs Go式(组合/接口) 对比
// 场景：一个日志系统中，需要支持多种输出目标（控制台、文件、网络）
// 以及多种格式化方式（纯文本、JSON）

package main

import (
    "fmt"
    "strings"
)

// =============================================
// Java 式：通过继承复用代码
// =============================================

// BaseLogger "基类"
type BaseLogger struct {
    Level string
}

func (b *BaseLogger) Log(msg string) {
    fmt.Printf("[%s] %s\n", b.Level, msg)
}

// ConsoleLogger "继承" BaseLogger
type ConsoleLogger struct {
    BaseLogger
    Prefix string
}

// FileLogger "继承" BaseLogger
type FileLogger struct {
    BaseLogger
    FilePath string
}

// NetworkLogger "继承" BaseLogger
type NetworkLogger struct {
    BaseLogger
    Endpoint string
}

// 问题：如果要新增 JSON 格式化，需要给每个 Logger 都加一个方法
// 或者再建一个继承层级 — 类数量爆炸

// =============================================
// Go 式：通过接口组合实现
// =============================================

// Output 输出目标接口 — 极小接口
type Output interface {
    Write(p []byte) (n int, err error)
}

// Formatter 格式化接口 — 极小接口
type Formatter interface {
    Format(level string, msg string) string
}

// Logger 组合接口
type Logger interface {
    Log(level string, msg string)
}

// 具体实现：Output 的实现
type ConsoleOutput struct {
    Prefix string
}

func (c *ConsoleOutput) Write(p []byte) (int, error) {
    fmt.Printf("%s%s", c.Prefix, string(p))
    return len(p), nil
}

type FileOutput struct {
    FilePath string
}

func (f *FileOutput) Write(p []byte) (int, error) {
    // 模拟写入文件
    fmt.Printf("[写入文件 %s] %s", f.FilePath, string(p))
    return len(p), nil
}

// 具体实现：Formatter 的实现
type TextFormatter struct{}

func (t *TextFormatter) Format(level string, msg string) string {
    return fmt.Sprintf("[%s] %s\n", level, msg)
}

type JSONFormatter struct{}

func (j *JSONFormatter) Format(level string, msg string) string {
    return fmt.Sprintf(`{"level":"%s","msg":"%s"}%s`, level, msg, "\n")
}

// 组合日志器：通过组合 Output + Formatter 实现 Logger
type ComposedLogger struct {
    Output    Output
    Formatter Formatter
}

func (l *ComposedLogger) Log(level string, msg string) {
    formatted := l.Formatter.Format(level, msg)
    l.Output.Write([]byte(formatted))
}

func main() {
    fmt.Println("=== Java 式：继承层级 ===")
    cl := &ConsoleLogger{BaseLogger: BaseLogger{Level: "INFO"}, Prefix: ">> "}
    cl.Log("hello") // 调用 BaseLogger.Log()

    fmt.Println("\n=== Go 式：接口组合 ===")
    // 控制台 + 纯文本
    l1 := &ComposedLogger{
        Output:    &ConsoleOutput{Prefix: ">> "},
        Formatter: &TextFormatter{},
    }
    l1.Log("INFO", "hello world")

    // 控制台 + JSON
    l2 := &ComposedLogger{
        Output:    &ConsoleOutput{Prefix: ">> "},
        Formatter: &JSONFormatter{},
    }
    l2.Log("ERROR", "something went wrong")

    // 文件 + JSON
    l3 := &ComposedLogger{
        Output:    &FileOutput{FilePath: "/var/log/app.log"},
        Formatter: &JSONFormatter{},
    }
    l3.Log("WARN", "disk space low")

    // 新增一种组合（比如网络输出 + 纯文本）只需要新建 Output 实现
    // 不需要修改任何已有代码 — 开闭原则

    // 统计组合数
    outputs := []string{"Console", "File", "Network"}
    formatters := []string{"Text", "JSON"}
    fmt.Printf("\n=== 可组合数：%d 种输出 × %d 种格式 = %d 种组合 ===\n",
        len(outputs), len(formatters), len(outputs)*len(formatters))
    fmt.Println("Go 式只需要实现", len(outputs), "个 Output +", len(formatters), "个 Formatter")
    fmt.Println("Java 式需要实现", len(outputs)*len(formatters), "个类（如果不考虑继承复用）")

    // 验证：所有格式都正确
    fmt.Println("\n=== 验证 ===")
    tests := []string{
        strings.TrimSpace((&TextFormatter{}).Format("INFO", "test")),
        strings.TrimSpace((&JSONFormatter{}).Format("ERROR", "test")),
    }
    for _, t := range tests {
        fmt.Printf("  Output: %q\n", t)
    }
}
