# dependency-file-anatomy

> 论据 E1：最小项目文件剖面  
> 目标：生成并比较 npm、Go Modules、Cargo、Python/pip lock 的依赖声明文件、锁定文件和校验文件。

## 环境

- Node.js v24.14.0 / npm 11.9.0
- Go 1.26.2 darwin/arm64
- Cargo 1.95.0 / rustc 1.95.0（Homebrew 安装）
- Python 3.9.6 / pip 25.3

## 子目录

| 子目录 | 声明文件 | 生成文件 | 验证命令 |
|--------|---------|---------|---------|
| npm-demo | package.json | package-lock.json | `npm --prefix npm-demo run verify` |
| go-demo | go.mod | go.sum | `go -C go-demo run .` |
| cargo-demo | Cargo.toml | Cargo.lock | `cargo run --manifest-path cargo-demo/Cargo.toml` |
| python-demo | requirements.in | pylock.toml | `python3 -m pip lock -r python-demo/requirements.in -o python-demo/pylock.toml` |

## 复现步骤

在本文项目根目录执行：

```bash
npm install --package-lock-only --ignore-scripts --prefix articles/package-manager-evolution/evidence/code/dependency-file-anatomy/npm-demo
npm install --ignore-scripts --prefix articles/package-manager-evolution/evidence/code/dependency-file-anatomy/npm-demo
npm --prefix articles/package-manager-evolution/evidence/code/dependency-file-anatomy/npm-demo run verify

go mod tidy -C articles/package-manager-evolution/evidence/code/dependency-file-anatomy/go-demo
go -C articles/package-manager-evolution/evidence/code/dependency-file-anatomy/go-demo run .
go -C articles/package-manager-evolution/evidence/code/dependency-file-anatomy/go-demo list -m all

cargo generate-lockfile --manifest-path articles/package-manager-evolution/evidence/code/dependency-file-anatomy/cargo-demo/Cargo.toml
cargo run --manifest-path articles/package-manager-evolution/evidence/code/dependency-file-anatomy/cargo-demo/Cargo.toml
cargo tree --manifest-path articles/package-manager-evolution/evidence/code/dependency-file-anatomy/cargo-demo/Cargo.toml

python3 -m pip lock -r articles/package-manager-evolution/evidence/code/dependency-file-anatomy/python-demo/requirements.in -o articles/package-manager-evolution/evidence/code/dependency-file-anatomy/python-demo/pylock.toml
```

## 输出文件

命令输出保存在：`../../output/dependency-file-anatomy/`。

统计分析保存在：`../../data/lockfile-anatomy.md`。
