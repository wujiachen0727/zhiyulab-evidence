# Go 抽象路径实验输出摘要

## 环境

- 标注：`[实测 Go 1.26.2 darwin/arm64]`
- 命令：
  - `go -C articles/generics-convergence/evidence/code/go-abstraction-paths run .`
  - `go -C articles/generics-convergence/evidence/code/go-abstraction-paths run ./compilefail`

## 关键结果

| 证据项 | 路径 | 结果 | 错误暴露阶段 |
|---|---|---|---|
| E1 | 复制粘贴 | `int`、`string` 各一套函数，输出正常 | 编译期能保类型，但新增类型需要复制函数 |
| E1 | `any/interface{}` | 函数可复用，但混入 `string` 后 `v.(int)` panic | 运行期 |
| E1 | 反射 | `[]UserID` 可去重，返回 `any`，调用者需相信运行期逻辑 | 运行期/调试期 |
| E2 | 泛型 | `[]UserID` 可去重且保留类型；`[]int{1, "2"}` 编译失败 | 编译期 |

## 原始输出摘录

### 正常运行

```text
[复制粘贴] int 去重: [1 2 3]
[复制粘贴] string 去重: [go java]
[泛型] UserID 去重: [u1 u2]
[any] 混入 string 后的暴露阶段: runtime panic -> true
[any] panic: interface conversion: interface {} is string, not int
[reflect] UserID 去重: [u1 u2] err: <nil>
```

### 泛型编译期失败

```text
compilefail/main.go:9:29: cannot use "2" (untyped string constant) as int value in array or slice literal
```

## 可用于正文的结论

泛型的核心收益不是“少写一个函数”，而是把错误暴露位置前移：`any/interface{}` 可以把函数体复用起来，但类型错误会混进运行路径；泛型让同类错误在编译期直接挡住。这就是“编译期门禁”隐喻可以落地的最小证据。
