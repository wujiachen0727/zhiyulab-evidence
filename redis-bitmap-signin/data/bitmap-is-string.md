# E1：Bitmap 底层是 String 的演示

## 测试方法

用 SETBIT 设置位，然后用 TYPE / OBJECT ENCODING / STRLEN / GET 查看 Bitmap 底层。

## 实测环境

Redis 8.8.0 / Go 1.26.4 / darwin/arm64

## 数据

```
SETBIT sign:demo 7 1
TYPE sign:demo        → string
OBJECT ENCODING sign:demo → raw
STRLEN sign:demo      → 1 字节

SETBIT sign:demo 100 1  (再设置第 100 位)
OBJECT ENCODING sign:demo → raw
STRLEN sign:demo      → 13 字节 (ceil(101/8) = 13)

GET sign:demo         → [1 0 0 0 0 0 0 0 0 0 0 0 8]
```

## 关键发现

1. TYPE 返回 `string`，证明 Bitmap 不是独立数据类型
2. OBJECT ENCODING 为 `raw`（String 的原始编码）
3. STRLEN 显示字节数，证明 SETBIT 操作的是 String 的字节
4. GET 能直接拿到字节序列：`[1 0 0 0 0 0 0 0 0 0 0 0 8]`
   - 第 0 字节 = 1（二进制 00000001，第 7 位被设置）
   - 第 12 字节 = 8（二进制 00001000，第 100 位 = 第 12 字节的第 4 位被设置）

## 结论

Bitmap 不是独立数据类型，底层就是 String。SETBIT/GETBIT/BITCOUNT 都是对 String 字节的位操作。

## 引用依据

Redis 官方文档原文："Bitmaps are not an actual data type, but a set of bit-oriented operations defined on the String type."
（https://redis.io/docs/latest/develop/data-types/bitmaps/）
