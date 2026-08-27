# E3+E4 泛型方法与接口 / reflect

Go 1.27.0，darwin/arm64。

## E3：方法自带类型参数不能实现接口

`interface_fail_method_tp.go` 编译失败：

```
Counter does not implement IntConverter (wrong type for method Convert)
    have Convert[U any](U) U
    want Convert(int) int
```

说明：限制针对**方法声明自己的类型参数**（如 `Convert[U any]`），不是「泛型类型上的普通方法」。

`Box[T].Value() T` 在 `T=int` 时可满足 `Value() int` 接口——**这不是本文要强调的边界**。

## E4：reflect 看不见带方法级类型参数的方法

`Counter` 有 `Plain()` 与 `Convert[U any](U) U`：

```
Counter NumMethod=1 names=[Plain]
```

`Convert` 未出现在方法集。库若依赖 reflect 枚举方法，会**漏掉**这类 API。

代码：`evidence/code/generic-method/`
