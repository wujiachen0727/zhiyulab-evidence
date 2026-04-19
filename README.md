# zhiyulab-evidence

> 止语Lab 技术文章配套的实验代码和实测数据——可信度承诺的底稿。

## 为什么有这个仓库

止语Lab 的长文都追求"论据自造"：能自己跑实验就不引用别人的结论。每一个实测数字、每一张 profile 都来自我在本地搭环境跑出来的原始文件。

这个仓库把这些实验代码和实测数据全部公开，目的有三：

1. **可复现**：读者可以亲自跑一遍，验证文章里的数字
2. **可质疑**：数据来源透明，谁都能翻看原始 pprof / trace / log 文件
3. **可继承**：实验代码本身可能比文章有更长的生命力——后来人可以在此基础上做自己的实验

## 目录结构

按文章 slug 分子目录，每篇文章一个独立空间：

```
zhiyulab-evidence/
└── go-profiling-toolchain/      # 《从 pprof 到持续 profiling：Go 性能工具链的三次升级》
    ├── e1-e4-sampling/          # 实验 1+4：pprof 100Hz 采样盲点
    ├── e2-wait-trap/            # 实验 2：pprof 给你一个数，trace 给你一个故事
    ├── e3-spike/                # 实验 3：时段毛刺被大窗口稀释
    └── e5-pyroscope-overhead/   # 实验 5：Pyroscope 全量 profile 开销实测
```

每个实验目录都有独立 README 说明复现步骤。

## 文章清单

| 文章 | 发布时间 | 子目录 | 配套实验数 |
|------|---------|--------|:---------:|
| [《从 pprof 到持续 profiling：Go 性能工具链的三次升级》](https://github.com/wujiachen0727/zhiyulab-evidence/tree/main/go-profiling-toolchain) | 2026-04 | `go-profiling-toolchain/` | 4 组（8 条独立论据）|

（后续文章发布时会在此追加）

## 复现原则

- **二进制不入库**：所有 `.go` 源码会入库，但编译产物（可执行文件）不入库。跑实验前请自己 `go build`
- **原始数据入库**：`.pprof`、`.trace`、`.csv`、`.log` 这类实测原始文件会保留
- **README 先读**：每个实验目录的 README 会写明环境要求（Go 版本、OS、依赖服务）

## 联系作者

- 作者：吴嘉晨 / 止语Lab
- 微信公众号：止语Lab
- 反馈：在对应文章下留言，或提 Issue

## License

MIT License（见 `LICENSE` 文件）
