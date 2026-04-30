# E8: JWT alg:none 漏洞 PoC

## 证伪结果

**假设**：JWT 手写 verify 不校验 alg 会放行 alg:none 攻击。  
**结果**：假设成立。PoC 演示攻击者构造 `alg:"none"` token，
不带签名就能通过 `verifyJWTUnsafe`，拿到任意 role（本例是 admin）。

## 运行

```bash
go run main.go
```

## 观察要点

- JWT 规范允许 `alg:none`，但业务绝不该接受
- 多数主流 JWT 库默认不接受 none，但**手写 verify 代码**很常见——
  尤其是多算法支持场景，开发者会在 switch 里加 `case "none": // 不验签`
- 核心防线：**白名单** —— 你接受哪几个 alg，其余一律拒绝。
  不要信任 header 里传进来的 alg 值来选择验签分支
- 历史案例：2015-2018 年间多个主流 JWT 库爆出 alg-confusion 漏洞（已修复）

## 环境

Go 1.26.2 darwin/arm64

## 对应章节

第三章 第 7 个幻觉 · 用了官方协议就安全
