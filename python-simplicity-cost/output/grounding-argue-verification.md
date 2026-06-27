# Grounding Argue 复核输出

执行日期：2026-05-29

## E1 类型检查实验复核

命令：

```bash
python3 -m mypy --version
python3 evidence/code/type-late-error/type_late_error_demo.py
python3 -m mypy evidence/code/type-late-error/type_late_error_demo.py
```

输出：

```text
mypy 1.19.1 (compiled: yes)
[实测 Python 3.9.6] 无类型检查路径：前两个输入正常，第三个输入到运行时才报错
case 1: ada
case 2: grace
case 3: AttributeError: 'NoneType' object has no attribute 'strip'
articles/python-simplicity-cost/evidence/code/type-late-error/type_late_error_demo.py:26: error: Incompatible types (expression has type "None", TypedDict item "name" has type "str")  [typeddict-item]
Found 1 error in 1 file (checked 1 source file)
```

## E2 GIL 并发边界实验复核

命令：

```bash
python3 evidence/code/gil-threading-demo/gil_threading_demo.py
```

输出：

```text
[实测 Python 3.9.6] CPU-bound 并发边界最小实验
platform=macOS-26.5-arm64-arm-64bit
executable=/Library/Developer/CommandLineTools/usr/bin/python3
cpu_count=14, tasks=4, n_per_task=3000000
sequential: elapsed=0.542s, speedup_vs_sequential=1.00x, checksum=576000968
threading-4: elapsed=0.604s, speedup_vs_sequential=0.90x, checksum=576000968
multiprocessing-4: elapsed=0.235s, speedup_vs_sequential=2.30x, checksum=576000968
```

## E3/E6 最小可交付项目烟测复核

命令：

```bash
cd evidence/code/script-to-project-cost/sample_workspace/03-deliverable-project
PYTHONPATH=src python3 -m unittest discover -s tests
python3 -m mypy src
```

输出：

```text
.
----------------------------------------------------------------------
Ran 1 test in 0.001s

OK
Success: no issues found in 3 source files
```
