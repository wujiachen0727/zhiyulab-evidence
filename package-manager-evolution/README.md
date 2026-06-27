# Evidence 索引：package-manager-evolution

> 文章：包管理器不是下载器，是构建信任的三层协议  
> 阶段：Argue / Step2 论据自造  
> 执行日期：2026-05-29

## 论据总览

| ID | 类型 | 论据 | 状态 | 路径 | 正文用途 |
|----|------|------|:----:|------|---------|
| E1 | 实验验证 | npm、Go、Cargo、Python 最小项目依赖文件剖面 | 完成 | `code/dependency-file-anatomy/` | 支撑“声明文件 ≠ 锁定/校验文件”和“四生态信任机制差异” |
| E2 | 场景模拟 | CI 依赖漂移排查路径 | 完成 | `scenarios/ci-dependency-drift.md` | 开头场景和工程规范章节 |
| E3 | 逻辑推演 | 三问矩阵：版本谁决定、来源谁证明、构建如何复现 | 完成 | `data/three-question-matrix.md` | 文章核心框架 |
| E4 | 数据实测 | 本地生成 lock/checksum 文件统计 | 完成 | `data/lockfile-anatomy.md` | 四生态剖面与对照表 |
| E5 | 经验落地 | 依赖升级 PR 审查规范 | 完成 | `data/dependency-review-practice.md` | 结尾行动清单 |

## 命令输出

| 生态 | 输出文件 | 说明 |
|------|---------|------|
| npm | `output/dependency-file-anatomy/npm-install.txt` | 生成 package-lock 并安装依赖 |
| npm | `output/dependency-file-anatomy/npm-verify.txt` | 运行样例，输出 `3 is odd: true` |
| Go | `output/dependency-file-anatomy/go-mod-tidy.txt` | 生成 go.sum |
| Go | `output/dependency-file-anatomy/go-run.txt` | 运行样例，输出 Go proverb |
| Go | `output/dependency-file-anatomy/go-list-m-all.txt` | 输出 build list |
| Cargo | `output/dependency-file-anatomy/cargo-generate-lockfile.txt` | 生成 Cargo.lock |
| Cargo | `output/dependency-file-anatomy/cargo-run.txt` | 运行样例，输出 `lock-diff-matters` |
| Cargo | `output/dependency-file-anatomy/cargo-tree.txt` | 输出依赖树 |
| Python | `output/dependency-file-anatomy/pip-lock.txt` | 生成 pylock.toml，包含 experimental 警告 |

## 自造度统计

- 独立论据：5 项
- 自造独立论据：5 项
- 外部引用：0 项作为核心论据；官方文档仅作为事实锚点
- 自造占比：100%

## 降级说明

无降级。Cargo 初始环境缺失，已按 practice-verify 规则通过 Homebrew 自动安装 Rust 工具链后完成实测。
