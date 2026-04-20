# Evidence 总索引

文章：别用 Go 写插件系统——但如果你非要写，这里有张决策表

## 论据清单

| ID | 类型 | 描述 | 状态 | 产出路径 |
|----|------|------|:----:|---------|
| E1 | 实验验证 | 5 方案 benchmark（原生/plugin接口/yaegi/wazero/IPC） | ✅ 完成 | `code/plugin-benchmark/` + `output/plugin-benchmark/` |
| E5 | 实验验证 | plugin.Open 重复加载返回旧符号 | ✅ 完成 | `code/plugin-reload/` + `output/plugin-reload/` |
| E2 | 场景模拟 | 3 种业务场景决策推理 | ✅ 完成 | `scenarios/three-scenarios.md` |
| E3 | 数据实测 | RPC 开销在不同频率下的体感影响 | ✅ 完成 | `data/rpc-overhead-impact.md` |
| E4 | 逻辑推演 | 从 Go 编译模型推导方案代价必然性 | ✅ 完成 | 融入正文 |

## 外部引用

| ID | 引用内容 | 原因 |
|----|---------|------|
| R1 | Go plugin 包官方文档 experimental 标注 | 证明官方态度 |
| R2 | go-plugin-benchmark (github.com/uberswe/go-plugin-benchmark) | 交叉验证自测数据 |

## 统计

- 自造论据：5 项（E1-E5）
- 外部引用：2 处（R1-R2）
- 总论据：7 项
- 自造占比：5/7 = **71%**（达标 ≥ 70%）
