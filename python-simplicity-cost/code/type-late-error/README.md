# type-late-error

用途：验证一类 Python 动态类型错误会延后到运行时暴露，而类型标注配合 mypy 可以把错误前移到检查期。

运行：

```bash
python3 type_late_error_demo.py
python3 -m mypy type_late_error_demo.py
```

说明：示例使用 `TypedDict` 描述输入结构。第三条输入把 `name` 写成 `None`，运行时只有执行到该输入才触发 `AttributeError`；mypy 会在静态检查阶段报告 `None` 与 `str` 不兼容。
