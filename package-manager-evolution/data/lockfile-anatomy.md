# Lockfile / 校验文件剖面统计

> 执行时间：2026-05-29  
> 标注：除“字段/包数量统计”为脚本读取本地生成文件所得，其余解释为基于文件内容的工程解读。

## 环境

| 生态 | 工具版本 | 样例依赖 | 生成文件 |
|------|---------|---------|---------|
| npm | npm 11.9.0 / Node v24.14.0 | is-odd@3.0.1 | package-lock.json |
| Go | go1.26.2 darwin/arm64 | rsc.io/quote v1.5.2 | go.mod + go.sum |
| Cargo | cargo 1.95.0 / rustc 1.95.0 | itertools 0.14.0 | Cargo.lock |
| Python | Python 3.9.6 / pip 25.3 | requests==2.32.5 | pylock.toml |

## 本地生成文件统计

| 生态 | 声明文件 | 锁定/校验文件 | 本地解析到的包/节点数 | 是否含来源 URL | 是否含 hash/checksum | 工程含义 |
|------|---------|--------------|:-------------------:|:-------------:|:------------------:|---------|
| npm | package.json | package-lock.json | 3 package entries | 是 | 是（integrity） | 锁文件记录完整 node_modules 解析结果，适合审查依赖树变化，但 diff 噪音较大 |
| Go | go.mod | go.sum | 4 build-list modules / 6 checksum lines | 间接来自 module proxy / module path | 是（go.sum） | 版本选择由 MVS 规则计算，go.sum 更像校验账本，不是传统依赖树快照 |
| Cargo | Cargo.toml | Cargo.lock | 3 package entries | 是（registry source） | 是（checksum） | Cargo.toml 保留范围，Cargo.lock 固化实际解析结果，构建系统心智更强 |
| Python | requirements.in | pylock.toml | 5 packages | 是（wheel url） | 是（5 wheel hashes） | pylock.toml 开始把 Python 环境从 requirements 文本推向可复现安装描述 |

## 三个可直接进入正文的观察

1. **声明文件和锁定文件不是同一种东西**：package.json/Cargo.toml/requirements.in/go.mod 更像“我想要什么”，lock/checksum 文件更像“这次实际拿到了什么”。
2. **Go 是关键反例**：Go 没有把完整依赖树快照进传统 lockfile，而是用 go.mod 的最小版本要求 + MVS + go.sum 校验形成确定性路径。
3. **可审查性是工程问题，不只是工具问题**：npm 的 package-lock 信息最完整但也最容易 noisy；Go 的 go.sum 更窄但 diff 更聚焦；Cargo.lock 和 pylock.toml 位于两者之间。

## 正文使用建议

- 用这张表支撑“快照层”和“协议层”的区别。
- 具体版本号可以写入正文，但不要过度泛化为生态永久特征。
- Python 部分要标注 pip lock/pylock.toml 当前仍属较新能力，避免夸大成熟度。
