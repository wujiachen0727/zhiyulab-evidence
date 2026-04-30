# E1: gosec 扫描 10 个主流 Go 开源项目

## 用途

扫描真实世界的 Go 代码，看看 gosec 发现的安全问题分布在哪——
用硬数据验证"第二层（开发者必守层）才是真高危所在"的论断。

## 运行方式

```bash
./run.sh          # 克隆 + 扫描（约 2-5 分钟）
python3 aggregate.py   # 聚合 + 打印分布
```

## 输入

10 个 Top Go 开源项目，来自各个领域：
Web 框架（gin/echo/mux）、CLI（cobra）、配置（viper）、测试（testify）、
日志（logrus）、数据库客户端（redis）、gRPC（grpc-go）、监控（client_golang）

## 输出

- `/tmp/gosec-results/gosec-{project}.json` — 每个项目的原始 gosec JSON 输出
- `evidence/data/gosec-distribution.json` — 聚合摘要
- `evidence/data/gosec-distribution.md` — 数据分析报告（人类可读）

## 核心结论

扫描了 ~40 万行 Go 代码（8 个项目成功解析），总发现 1241 个 issues：

- 代码卫生（未处理错误 G104）：67.0%
- 开发者必守层：22.8%
- 兜底层相关（unsafe/cgi/range 指针）：10.2%

**但按 HIGH 严重度算：100% 的 HIGH issues 都在开发者必守层。**

这验证了本文核心观点：语言兜底层很少出严重问题（Go 设计就是这样），
真正高危的全部是开发者必须自己守的 API 层问题。

## 可复现性

- gosec 版本：dev @ 2026-04
- Go 版本：1.26.2
- OS：darwin/arm64
- 项目 HEAD：克隆时的最新 main 分支（`--depth 1`）

不同版本 gosec 规则可能有微小差异（G115、G123 等是较新加入的规则），
但按严重度的大盘分布结论稳健。
