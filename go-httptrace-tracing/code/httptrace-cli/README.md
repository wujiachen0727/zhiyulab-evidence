# httptrace CLI 工具

## 环境

- Go 1.26.4 darwin/arm64
- macOS (Darwin)

## 运行说明

```bash
go run main.go <url>
```

示例：
```bash
go run main.go https://github.com
go run main.go https://baidu.com
```

## 输出说明

工具会打印 HTTP 请求各阶段的耗时和连接复用信息。
