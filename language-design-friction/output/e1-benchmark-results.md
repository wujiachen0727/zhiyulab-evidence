# E1: Go reflect vs generics benchmark 结果

## 环境

- Go 1.26.2 darwin/arm64
- Apple M4 Pro
- 2026-05-26 实测

## 原始数据（3 次运行取平均）

| 方法 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| DirectAccess（基线） | 1.7 | 0 | 0 |
| GetFieldGeneric | 3.2 | 0 | 0 |
| ReflectTypeCheck（仅类型检查） | 1.8 | 0 | 0 |
| ReflectTraverseAll（遍历所有字段） | 39.1 | 80 | 1 |
| GetFieldReflect（完整安全检查） | 73.7 | 85 | 2 |

## 关键洞察

1. **性能差距**：reflect 完整调用 vs 泛型等价操作 = **22x 慢**
2. **分配差距**：reflect 每次调用 2 allocs，泛型 0 allocs
3. **层次递进**：
   - 类型检查本身很快（1.8 ns）
   - 但一旦访问 field 值 + Interface() 转换 → 触发堆分配
   - 安全检查层（CanInterface/FieldByName 等）叠加后代价陡增
4. **代码行数差距**：reflect 方案 ~20 行（含 3 层防御检查） vs 泛型 ~3 行

## 正文可引用的表述

- "同一任务，reflect 版本比泛型版本慢 22 倍，多 2 次堆分配"
- "Go reflect API 的每一层检查（Kind → FieldByName → CanInterface）都是设计者故意加的'你确定？'"
- "仅仅是 reflect.ValueOf 加一个类型检查只要 1.8 ns，但完整的'安全使用'路径需要 73 ns——额外的 71 ns 就是'认知税'"
