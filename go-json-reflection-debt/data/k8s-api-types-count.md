# K8s API Types 数量统计

> 数据来源：github.com/kubernetes/api（2026-05-31 最新 master，shallow clone）
> 标注：[实测]

## 统计结果

- **types.go 文件数**：58 个
- **struct 定义总数**：1101 个
- **最大单文件**：core/v1/types.go 含 240 个 struct

## Top 10 struct 密集文件

| 文件 | struct 数量 |
|------|:-----------:|
| core/v1/types.go | 240 |
| resource/v1beta2/types.go | 49 |
| extensions/v1beta1/types.go | 45 |
| resource/v1beta1/types.go | 44 |
| resource/v1/types.go | 44 |
| admissionregistration/v1/types.go | 36 |
| networking/v1/types.go | 35 |
| admissionregistration/v1beta1/types.go | 34 |
| apps/v1beta2/types.go | 33 |
| apps/v1/types.go | 30 |

## Codegen 推演

如果 Go 标准库选择 codegen 路（类似 easyjson），对 Kubernetes 意味着：

1. **生成文件数量**：至少 58 个 `*_easyjson.go` 文件
2. **生成代码行数**：按 easyjson 平均每 struct ~100 行 → 约 110,000 行生成代码
3. **构建时间增加**：每次 `go generate` 需要解析 1101 个 struct 的 reflect 信息
4. **维护成本**：每次修改 struct 字段（K8s 每个版本都有大量字段变动）→ 重新 generate → review 生成代码的 diff → CI 检查生成代码是否过时

**结论**：对 K8s 这种规模的项目，codegen 路意味着把"运行时反射的 CPU 代价"换成"开发时的维护代价 + 编译时间代价 + 代码膨胀"。Go 团队选择标准库不走 codegen，一个重要考量就是"不把这种维护成本推给所有用户"。
