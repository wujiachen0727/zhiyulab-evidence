# E1 类型错误前移实验结果

[实测 Python 3.9.6 + mypy 1.19.1]

## 核心假设与证伪点

- 假设：至少一类 Python 错误可以从运行时前移到检查期。
- 证伪点：如果 mypy 无法发现 `TypedDict` 字段类型错误，则改用更典型的 Optional / 容器元素案例。
- 结果：证伪未发生。mypy 在检查期发现 `name` 字段的 `None` 与 `str` 不兼容。

## 方法

构造同一组输入：前两条 `name` 是字符串，第三条 `name` 是 `None`。无类型检查路径直接运行；类型标注路径使用 `TypedDict` 声明 `name: str`，再用 mypy 检查。

## 关键输出

运行时输出：

```text
[实测 Python 3.9.6] 无类型检查路径：前两个输入正常，第三个输入到运行时才报错
case 1: ada
case 2: grace
case 3: AttributeError: 'NoneType' object has no attribute 'strip'
```

mypy 输出：

```text
articles/python-simplicity-cost/evidence/code/type-late-error/type_late_error_demo.py:26: error: Incompatible types (expression has type "None", TypedDict item "name" has type "str")  [typeddict-item]
Found 1 error in 1 file (checked 1 source file)
```

## 可供正文引用的结论

同一个 `None` 输入，在无类型检查路径里要等到第三条数据执行到 `.strip()` 才爆炸；加入 `TypedDict` 后，mypy 在代码运行前就能指出字段类型不兼容。这里的账不是消失了，而是从运行时排障转移到了类型标注和检查工具上。
