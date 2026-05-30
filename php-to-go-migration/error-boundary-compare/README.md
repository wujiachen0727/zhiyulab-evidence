# error-boundary-compare

## 目标

比较同一组边界输入在 PHP 与 Go 中暴露错误的位置。

输入：

```json
{"user_id":"42","email":123}
```

观察点：

- PHP 默认弱类型函数参数是否会进行类型转换。
- PHP `strict_types=1` 是否会在函数调用边界抛出类型错误。
- Go `encoding/json` 解码到强类型 struct 时是否在解码边界返回错误。

## 运行方式

```bash
bash run.sh
```

## 环境

- Go：本机 `go version`
- PHP：Docker 镜像 `php:8.4-cli`

## 解释边界

这个实验不能证明“PHP 做不到类型前移”。它恰好要验证反例：现代 PHP 通过 `strict_types=1` 也能把一部分错误前移到函数调用边界。本文真正要讨论的是默认工程习惯和复杂度归属差异，而不是语言能力优劣。
