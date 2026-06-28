# E4: THP 放大效应实验

## 实验目的

对比 THP 开启/关闭时 COW 内存复制量，验证 THP 对 fork/COW 的放大效应。

## 运行方式

```bash
cd evidence/code
bash run-all-experiments.sh
```

结果输出到 `evidence/output/e1-e6-results.md` § E4。

## 降级说明

macOS Docker Desktop (LinuxKit) 环境下实测结果与理论预期相反（THP OFF 的 COW 反而更大）。改为逻辑推演 + Netdata 案例引用。

**标注**：[推演 + 外部引用] 非本环境实测。
