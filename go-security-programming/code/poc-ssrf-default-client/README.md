# E6: http.Client 默认信任任何 URL（SSRF）

## 证伪结果

**假设**：`http.Get(userURL)` 没有任何信任边界。  
**结果**：假设成立。默认 http.Client 对 loopback、私有段、link-local 地址一视同仁。
PoC 演示了抓取 loopback 上的"伪 metadata"服务并泄漏 AWS credentials。

## 运行

```bash
go run main.go
```

## 观察要点

- 真实威胁目标：AWS/GCP 云 metadata (169.254.169.254)、内网 Redis/ES 管理接口
- 防御要点：IP 白名单 + 自定义 DialContext + 禁止自动跟随重定向 + 处理 DNS rebinding
- 标准库不会帮你做，你必须自己写

## 环境

Go 1.26.2 darwin/arm64

## 对应章节

第三章 第 5 个幻觉 · http.Client 默认就够安全
