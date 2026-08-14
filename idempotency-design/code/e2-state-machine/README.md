# E2 状态机防乱序实验

## 运行环境

- Go 1.26.2 (darwin/arm64)
- 无外部依赖，纯标准库

## 运行方式

```bash
cd e2-state-machine
go run main.go
```

## 实验设计

模拟支付回调乱序到达（业务顺序：先 paid 后 shipped，实际网络先收到 shipped）：

| 组 | 说明 | 预期 |
|----|------|------|
| 对照组 A | 无状态机，回调直接改状态 | 发货回调先到 → 状态变 shipped；支付回调后到 → 状态变 paid（状态来回跳） |
| 实验组 B | 有状态机（pending→paid→shipped） | 发货回调被拒（尚未支付不能发货）；支付回调合法迁移 |

## 结果摘要

```
对照组A: 收到发货回调→shipped: pending → shipped
         收到支付成功回调→paid: shipped → paid
         >>> 中间出现过'未支付却已发货'的错误状态
实验组B: 收到发货回调→shipped: pending → shipped ❌ 非法迁移被拒
         收到支付成功回调→paid: pending → paid ✅
场景3：重复 paid 回调 → 第1次生效，第2/3次被忽略（幂等）
```

## 结论

- 状态机拒绝非法迁移，乱序回调不会造成错误状态
- 同状态重复回调天然幂等（已处于 paid 的 paid 回调被忽略）
- 完整输出见 `../../output/e2-state-machine/result.txt`
