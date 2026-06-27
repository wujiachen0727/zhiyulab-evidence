# 论据总索引

**文章**：从panic到优雅降级：Go错误分层的实战指南
**更新日期**：2026-04-11

## 已完成论据

| ID | 类型 | 描述 | 状态 | 产出路径 |
|----|------|------|:----:|---------|
| E1 | 实验验证 | 三层分层对比实验：无分层暴露5项敏感信息，有分层零泄露 | ✅ 完成 | evidence/code/error-layering-compare/ + evidence/output/error-layering-compare/ |
| E5 | 实验验证 | errors.Join实测：errors.Is可穿透组合错误，字符串拼接无法做到 | ✅ 完成 | evidence/code/errors-join-demo/ |
| E6 | 数据实测 | panic+recover vs error return性能对比：panic慢670倍，2 allocs vs 0 allocs | ✅ 完成 | evidence/code/panic-vs-error-bench/ |
| E3 | 场景模拟 | DB超时→错误泄露→攻击链路场景 | 📝 直接融入正文 | — |
| E4 | 场景模拟 | 用户踩坑经历（降级：无直接经历→场景模拟） | 📝 直接融入正文 | — |
| E2 | 数据实测 | 5个开源Go项目panic滥用统计 | ⏳ 执行中 | evidence/data/panic-abuse-stats.md |

## 表达手法

| ID | 类型 | 描述 | 说明 |
|----|------|------|------|
| E7 | 逻辑推演 | Go官方关闭语法提案→"问题不在语法"的推导 | 串联核心论点 |
| E8 | 类比桥接 | 错误分层 ≈ 网络分层 | 辅助理解三层架构 |

## 外部引用

| ID | 引用内容 | 引用来源 |
|----|---------|---------|
| R1 | Go官方博客 error-syntax 核心结论 | go.dev/blog/error-syntax |
| R2 | CVE-2025-7445 | CVE数据库 |
