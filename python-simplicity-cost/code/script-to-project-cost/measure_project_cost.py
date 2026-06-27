from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import shutil
import textwrap


ROOT = Path(__file__).resolve().parent
WORKSPACE = ROOT / "sample_workspace"
DATA_DIR = ROOT.parents[1] / "data"
OUTPUT_DIR = ROOT.parents[1] / "output" / "script-to-project-cost"


@dataclass(frozen=True)
class Stage:
    name: str
    files: dict[str, str]
    commands: list[str]
    config_items: list[str]
    delivery_checks: list[str]


STAGES = [
    Stage(
        name="00-single-script",
        files={
            "invoice_summary.py": """
                import csv
                from pathlib import Path

                total = 0
                with Path('invoices.csv').open() as file:
                    for row in csv.DictReader(file):
                        total += int(row['amount'])
                print(total)
            """,
        },
        commands=["python invoice_summary.py"],
        config_items=[],
        delivery_checks=["能在作者机器上运行"],
    ),
    Stage(
        name="01-repeatable-script",
        files={
            "invoice_summary.py": """
                import csv
                from pathlib import Path

                def summarize(path: str) -> int:
                    total = 0
                    with Path(path).open() as file:
                        for row in csv.DictReader(file):
                            total += int(row['amount'])
                    return total

                if __name__ == '__main__':
                    print(summarize('invoices.csv'))
            """,
            "README.md": "# invoice summary\n\nRun: python invoice_summary.py\n",
            "requirements.txt": "# stdlib only\n",
        },
        commands=["python -m venv .venv", "python invoice_summary.py"],
        config_items=["requirements.txt"],
        delivery_checks=["运行说明", "依赖口径", "可复跑命令"],
    ),
    Stage(
        name="02-testable-project",
        files={
            "src/invoice_summary/__init__.py": "",
            "src/invoice_summary/core.py": """
                import csv
                from pathlib import Path

                def summarize(path: str) -> int:
                    total = 0
                    with Path(path).open() as file:
                        for row in csv.DictReader(file):
                            total += int(row['amount'])
                    return total
            """,
            "tests/test_core.py": """
                import csv
                from pathlib import Path
                import tempfile
                import unittest

                from invoice_summary.core import summarize

                class SummaryTest(unittest.TestCase):
                    def test_summarize_amounts(self):
                        with tempfile.TemporaryDirectory() as tmp:
                            path = Path(tmp) / 'invoices.csv'
                            with path.open('w', newline='') as file:
                                writer = csv.DictWriter(file, fieldnames=['amount'])
                                writer.writeheader()
                                writer.writerows([{'amount': '10'}, {'amount': '20'}])
                            self.assertEqual(summarize(str(path)), 30)

                if __name__ == '__main__':
                    unittest.main()
            """,
            "README.md": "# invoice summary\n\nRun tests with PYTHONPATH=src python -m unittest discover -s tests\n",
            "requirements.txt": "# stdlib only\n",
        },
        commands=["PYTHONPATH=src python -m unittest discover -s tests", "python -m venv .venv", "python -m pip install -r requirements.txt"],
        config_items=["src layout", "tests", "requirements.txt"],
        delivery_checks=["单元测试", "包结构", "依赖安装命令", "运行说明"],
    ),
    Stage(
        name="03-deliverable-project",
        files={
            "src/invoice_summary/__init__.py": "",
            "src/invoice_summary/core.py": """
                import csv
                from pathlib import Path

                def summarize(path: str) -> int:
                    total = 0
                    with Path(path).open() as file:
                        for row in csv.DictReader(file):
                            total += int(row['amount'])
                    return total
            """,
            "src/invoice_summary/__main__.py": """
                from .core import summarize

                if __name__ == '__main__':
                    print(summarize('invoices.csv'))
            """,
            "tests/test_core.py": """
                import csv
                from pathlib import Path
                import tempfile
                import unittest

                from invoice_summary.core import summarize

                class SummaryTest(unittest.TestCase):
                    def test_summarize_amounts(self):
                        with tempfile.TemporaryDirectory() as tmp:
                            path = Path(tmp) / 'invoices.csv'
                            with path.open('w', newline='') as file:
                                writer = csv.DictWriter(file, fieldnames=['amount'])
                                writer.writeheader()
                                writer.writerows([{'amount': '10'}, {'amount': '20'}])
                            self.assertEqual(summarize(str(path)), 30)
            """,
            "pyproject.toml": """
                [project]
                name = "invoice-summary"
                version = "0.1.0"
                requires-python = ">=3.9"

                [tool.mypy]
                python_version = "3.9"
                strict = true
            """,
            "README.md": "# invoice summary\n\nRun: python -m invoice_summary\nTest: PYTHONPATH=src python -m unittest discover -s tests\nType check: python -m mypy src\n",
            "Makefile": "test:\n\tPYTHONPATH=src python -m unittest discover -s tests\n\ntypecheck:\n\tpython -m mypy src\n",
            ".github/workflows/ci.yml": """
                name: ci
                on: [push]
                jobs:
                  test:
                    runs-on: ubuntu-latest
                    steps:
                      - uses: actions/checkout@v4
                      - uses: actions/setup-python@v5
                        with:
                          python-version: '3.9'
                      - run: PYTHONPATH=src python -m unittest discover -s tests
            """,
        },
        commands=["python -m venv .venv", "python -m pip install -e .", "PYTHONPATH=src python -m unittest discover -s tests", "python -m mypy src", "make test"],
        config_items=["pyproject.toml", "mypy strict", "Makefile", "CI workflow", "src layout", "tests"],
        delivery_checks=["可安装", "可测试", "可类型检查", "可 CI 复跑", "有 README", "有运行入口"],
    ),
]


def write_files(stage_dir: Path, files: dict[str, str]) -> None:
    for relative, content in files.items():
        path = stage_dir / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(textwrap.dedent(content).strip() + "\n")


def count_files(stage_dir: Path) -> int:
    return sum(1 for path in stage_dir.rglob("*") if path.is_file())


def main() -> None:
    if WORKSPACE.exists():
        shutil.rmtree(WORKSPACE)
    WORKSPACE.mkdir(parents=True)
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    rows = []
    for stage in STAGES:
        stage_dir = WORKSPACE / stage.name
        write_files(stage_dir, stage.files)
        rows.append(
            {
                "stage": stage.name,
                "file_count": count_files(stage_dir),
                "command_count": len(stage.commands),
                "config_count": len(stage.config_items),
                "delivery_check_count": len(stage.delivery_checks),
                "commands": "; ".join(stage.commands),
                "config_items": "; ".join(stage.config_items) if stage.config_items else "无",
                "delivery_checks": "; ".join(stage.delivery_checks),
            }
        )

    csv_lines = ["stage,file_count,command_count,config_count,delivery_check_count"]
    for row in rows:
        csv_lines.append(
            f"{row['stage']},{row['file_count']},{row['command_count']},{row['config_count']},{row['delivery_check_count']}"
        )
    (OUTPUT_DIR / "cost-counts.csv").write_text("\n".join(csv_lines) + "\n")

    md = [
        "# 脚本到可交付项目的复杂度账单",
        "",
        "[实测 Python 3.9.6] 计数口径：文件数、命令数、配置项数、交付检查项数。示例项目不连接外部服务，只统计从个人脚本到可复跑/可测试/可交付所需显式约束。",
        "",
        "| 阶段 | 文件数 | 命令数 | 配置项数 | 交付检查项数 |",
        "|------|------:|------:|----------:|--------------:|",
    ]
    for row in rows:
        md.append(
            f"| {row['stage']} | {row['file_count']} | {row['command_count']} | {row['config_count']} | {row['delivery_check_count']} |"
        )
    md.extend(["", "## 明细", ""])
    for row in rows:
        md.extend(
            [
                f"### {row['stage']}",
                f"- 命令：{row['commands']}",
                f"- 配置项：{row['config_items']}",
                f"- 交付检查项：{row['delivery_checks']}",
                "",
            ]
        )
    md.extend(
        [
            "## 可供正文引用的结论",
            "",
            "从单文件脚本到最小可交付项目，文件数从 1 增加到 8，显式命令从 1 增加到 5，配置项从 0 增加到 6。这个结果不说明 Python ‘不好’，只说明脚本层省掉的工程约束，在交付层会重新出现。",
        ]
    )
    (DATA_DIR / "script-to-project-cost.md").write_text("\n".join(md) + "\n")


if __name__ == "__main__":
    main()
