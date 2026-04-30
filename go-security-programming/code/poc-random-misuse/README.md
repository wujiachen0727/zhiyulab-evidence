# E3: math/rand vs crypto/rand 误用

## 证伪结果

**假设**：math/rand 生成的 token 不是密码学安全。  
**结果**：假设成立。math/rand/v2 默认用系统随机种子降低了部分风险，
但算法（PCG）仍不是密码学安全的——只要种子泄漏或内部状态被恢复，输出可预测。

## 运行

```bash
go run main.go
```

## 环境

Go 1.26.2 darwin/arm64

## 对应章节

第三章 第 2 个幻觉 · math/rand 看起来像 crypto/rand
