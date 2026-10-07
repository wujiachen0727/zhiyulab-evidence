# typed-nil-interface 运行结果

**对应题目**：三、接口里的 nil
**目的**：确认"接口非 nil 但值是 nil"的判定，以及这种错误类型在调用方的实际后果
**运行方式**：`cd code/typed-nil-interface && go run .`
**原始输出**：`result.txt`（stdout）

## 关键输出

```
runGood()  == nil            -> true
runSafe()  == nil            -> false
runSafe()  的动态类型         -> *main.SafeErr
runSafe()  的动态值           -> boom(未初始化)

调用方判定为出错，错误消息："boom(未初始化)"

调用方判定为出错，接着调用 err.Error() 打日志……
err.Error() 触发 panic：runtime error: invalid memory address or nil pointer dereference

errors.As 取出后 target == nil，可以判定这是个空指针错误
```

## 结论

接口值只有"类型和值都为 nil"时才等于 nil。返回 typed nil 指针会让调用方走进错误分支，把一条不存在的错误记进日志。

更糟的一种：如果这个错误类型的 `Error()` 方法需要读自己的字段，nil 接收者一调用就 panic——错误处理分支里的日志语句成了压垮进程的那根稻草。

识别方式：用 `errors.As` 取出具体类型，再判断取出的指针是否为 nil。
