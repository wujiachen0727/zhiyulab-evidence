# 论据总索引

> 生成时间：2026-06-08
> 论证阶段：Step 2 论据自造

## 论据清单

| # | 类型 | 名称 | 代码路径 | 输出路径 | 状态 | 使用章节 |
|---|------|------|---------|---------|:----:|:-------:|
| E1 | 经验落地 | Java in Go 模拟继承的陷阱 | `evidence/code/e1-java-in-go/` | `evidence/output/e1-results.md` | ✅ | 开场、第一章 |
| E2 | 实验验证 | 日志系统的 Java 式(继承) vs Go 式(接口组合) | `evidence/code/e2-java-vs-go/` | `evidence/output/e2-results.md` | ✅ | 开场、第一章 |
| E3 | 场景模拟 | HTTP 中间件链的自然演化 | `evidence/code/e3-middleware-evolution/` | `evidence/output/e3-results.md` | ✅ | 第二章、第四章 |
| E4 | 经验落地 | Go 标准库中的设计模式实例 | `evidence/code/e4-stdlib-patterns/` | `evidence/output/e4-results.md` | ✅ | 第一章、第二章、第三章 |
| E5 | 数据实测 | Functional Options vs Builder 模式 | `evidence/code/e5-options-vs-builder/` | `evidence/output/e5-results.md` | ✅ | 第三章 |
| E6 | 逻辑推演 | 从 Go 设计哲学推导模式选择方向 | 融入正文 | 融入正文 | ✅ | 各章过渡 |

## 自造度统计

- 自造论据：6 条（E1-E6）
- 外部引用：2 处（Go Proverbs、Tony Bai 七宗罪）
- 自造占比：6 / (6 + 2) = **75%** ✅ ≥ 70%

## 论据使用计划

| 章节 | 使用论据 |
|------|---------|
| 开场（"Java in Go"代码） | E1（继承陷阱）、E2（对比代码）|
| 第一章（组合优于继承） | E1（继承陷阱）、E2（对比代码）、E4（标准库实例）|
| 第二章（接口精简） | E3（中间件演化）、E4（http.Handler）|
| 第三章（零值有用/工厂） | E4（sync.Once）、E5（Options vs Builder）|
| 第四章（并发原生） | E3（channel 中间件）、E4（context 观察者）|
