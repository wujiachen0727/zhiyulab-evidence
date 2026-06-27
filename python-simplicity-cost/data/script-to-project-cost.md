# 脚本到可交付项目的复杂度账单

[实测 Python 3.9.6] 计数口径：文件数、命令数、配置项数、交付检查项数。示例项目不连接外部服务，只统计从个人脚本到可复跑/可测试/可交付所需显式约束。

| 阶段 | 文件数 | 命令数 | 配置项数 | 交付检查项数 |
|------|------:|------:|----------:|--------------:|
| 00-single-script | 1 | 1 | 0 | 1 |
| 01-repeatable-script | 3 | 2 | 1 | 3 |
| 02-testable-project | 5 | 3 | 3 | 4 |
| 03-deliverable-project | 8 | 5 | 6 | 6 |

## 明细

### 00-single-script
- 命令：python invoice_summary.py
- 配置项：无
- 交付检查项：能在作者机器上运行

### 01-repeatable-script
- 命令：python -m venv .venv; python invoice_summary.py
- 配置项：requirements.txt
- 交付检查项：运行说明; 依赖口径; 可复跑命令

### 02-testable-project
- 命令：PYTHONPATH=src python -m unittest discover -s tests; python -m venv .venv; python -m pip install -r requirements.txt
- 配置项：src layout; tests; requirements.txt
- 交付检查项：单元测试; 包结构; 依赖安装命令; 运行说明

### 03-deliverable-project
- 命令：python -m venv .venv; python -m pip install -e .; PYTHONPATH=src python -m unittest discover -s tests; python -m mypy src; make test
- 配置项：pyproject.toml; mypy strict; Makefile; CI workflow; src layout; tests
- 交付检查项：可安装; 可测试; 可类型检查; 可 CI 复跑; 有 README; 有运行入口

## 可供正文引用的结论

从单文件脚本到最小可交付项目，文件数从 1 增加到 8，显式命令从 1 增加到 5，配置项从 0 增加到 6。这个结果不说明 Python ‘不好’，只说明脚本层省掉的工程约束，在交付层会重新出现。
