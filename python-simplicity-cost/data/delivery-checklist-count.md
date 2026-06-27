# 可交付检查清单数量统计

[实测 Python 3.9.6] 统计对象：`evidence/code/script-to-project-cost/sample_workspace/03-deliverable-project/`。  
统计口径：只统计从“能在作者机器上运行”升级到“别人能复跑、能改、能交付”所需的显式检查项，不把代码行数当作复杂度指标。

## 检查项

| # | 检查项 | 对应产物 | 为什么需要 |
|---|--------|----------|------------|
| 1 | 运行入口 | `src/invoice_summary/__main__.py` | 别人不需要知道内部函数名也能运行 |
| 2 | 包结构 | `src/invoice_summary/` | 从单文件脚本变成可导入、可测试的模块 |
| 3 | 自动测试 | `tests/test_core.py` | 改逻辑后能判断是否破坏旧行为 |
| 4 | 类型检查配置 | `pyproject.toml [tool.mypy]` | 把一部分错误前移到检查期 |
| 5 | 项目元信息 | `pyproject.toml [project]` | 声明名称、版本、Python 版本边界 |
| 6 | 运行说明 | `README.md` | 让非作者知道怎么跑、怎么测、怎么检查 |
| 7 | 统一命令 | `Makefile` | 把常用命令固化，减少口口相传 |
| 8 | CI 入口 | `.github/workflows/ci.yml` | 让检查在协作边界自动复跑 |

## 烟测结果

```text
.
----------------------------------------------------------------------
Ran 1 test in 0.001s

OK
Success: no issues found in 3 source files
```

## 可供正文引用的结论

“能跑”的标准只有 1 条：作者机器上执行成功。最小“可交付”的标准至少变成 8 条：运行入口、包结构、测试、类型检查、项目元信息、运行说明、统一命令和 CI 入口。Python 没有把这些复杂度制造出来，但它也不会替你自动消灭这些复杂度。
