# E7：签到场景完整实现

## 功能

- 每日签到（SETBIT）
- 查询某天是否签到（GETBIT）
- 本月签到次数（BITCOUNT）
- 连续签到天数（BITFIELD GET + 位运算）

## key 设计

```
sign:{uid}:{yyyyMM}
```

每月一个 Bitmap，offset = day-1（第 1 天对应 offset 0）。

## 运行方式

```bash
cd evidence/code/signin-implementation
go mod init signin-implementation
go get github.com/redis/go-redis/v9
go run main.go
```

## 预期输出

见 `../../output/signin-implementation/result.txt`

## 关键点

1. **连续签到天数用 BITFIELD GET**：一次命令读取整个月的位图，避免逐位 GETBIT 的 N 次网络往返
2. **BITFIELD 位序**：offset 0 = 最高位（大端序），所以当天（offset=day-1）对应返回值的最低位
3. **跨月查询**：本实现只处理单月内，跨月需调用方按月分段查询后拼接
