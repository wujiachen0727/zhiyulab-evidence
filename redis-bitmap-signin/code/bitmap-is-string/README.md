# E1：Bitmap 底层是 String

## 运行方式

```bash
cd evidence/code/bitmap-is-string
go mod init bitmap-is-string
go get github.com/redis/go-redis/v9
go run main.go
```

## 预期输出

见 `../../output/bitmap-is-string/result.txt`

## 结论

Bitmap 不是独立数据类型，底层是 String。SETBIT 是对 String 字节的位操作。
