# keepalive-matrix

## 用途

生成 TCP keepalive、连接池 idle timeout 与 NAT/LB idle timeout 的配置组合对比表，用于支撑文章中的“探测早于回收，失败早于用户感知”原则。

## 运行环境

- Python 3.9+
- 无第三方依赖

## 运行方式

```bash
python3 analyze_keepalive.py
```

## 输出

输出 Markdown 表格，保存到：

- `evidence/output/keepalive-matrix/result.md`

## 证据性质

- 配置组合表为 `[推演]`，不是生产环境实测。
- Linux 默认 keepalive 参数来自 `tcp(7)` 文档事实。
- AWS/Azure idle timeout 来自立意阶段求证快照。
