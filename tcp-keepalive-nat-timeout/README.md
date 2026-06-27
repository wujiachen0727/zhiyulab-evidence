# Evidence 论据自造记录

> 由 Practice-Verify 阶段维护。外部引用是最后手段，自造论据优先。

## 自造论据清单

| ID | 类型 | 文件/目录 | 方法 | 结论 | 可复现 |
|---|---|---|---|---|:---:|
| E1 | 经验落地 / 边界声明 | `evidence/scenarios/experience-boundary.md` | 基于用户已确认的排查主题，记录可写与不可写边界 | 当前缺少生产日志/抓包/配置，不能编造具体现场；正文只使用排查故事骨架 | 否 |
| E2 | 数据实测 | `evidence/data/local-tcp-keepalive-snapshot.md` | 本机执行 `sysctl net.inet.tcp.keepidle net.inet.tcp.keepintvl net.inet.tcp.keepcnt` | 本机 macOS 出现 2h 首次探测、75s 间隔；但 keepcnt 与 Linux 默认不同，正文需区分平台 | 是 |
| E3 | 场景模拟 | `evidence/scenarios/connection-lifecycle-timeline.md` | 构造 T=0、T≈300s、T=7200s 的连接生命周期 | 中间设备回收状态时，应用层不一定立刻知道；下一次复用旧连接才暴露异常 | 是 |
| E4 | 逻辑推演 | `evidence/scenarios/connection-lifecycle-timeline.md` | 明确前提：keepalive 首次探测晚于 idle timeout，且应用/连接池没有更早回收 | 当条件成立时，TCP 探测来得太晚，无法阻止陈旧连接窗口 | 是 |
| E5 | 数据实测 / 配置推演 | `evidence/code/keepalive-matrix/`、`evidence/output/keepalive-matrix/result.md` | Python 脚本生成不同 keepalive / idle timeout / 连接池组合对比表 | 修复原则不是固定数字，而是主动探测或连接池回收早于中间设备回收 | 是 |
| E6 | 场景模拟 | `evidence/scenarios/troubleshooting-decision-table.md` | 把超时、RST、重传、复用连接失败等信号整理成排查决策表 | 如果失败集中在长时间空闲后的首次复用，应从旧连接和中间设备回收窗口查起 | 是 |

## 外部引用清单

| ID | 来源 | 用途 | 是否必须 |
|---|---|---|:---:|
| R1 | Linux `tcp(7)` manual page | 支撑 Linux keepalive 默认参数含义与默认值 | 是 |
| R2 | AWS NAT Gateway troubleshooting；Microsoft Learn Azure Load Balancer | 支撑分钟级 NAT/LB idle timeout 的现实存在，并提醒不同基础设施不同 | 是 |

## 比例统计

- 完整自造论据数：5（E2-E6）
- 部分完成论据数：1（E1，现场细节不足，仅保留边界与叙事骨架）
- 外部引用数：2（R1-R2）
- 自造比例（按完整自造计）：5 / (5 + 2) = 71.4%
- 自造比例（含 E1 边界论据）：6 / (6 + 2) = 75%
- 判定：达到 ≥70% 目标；但 E1 现场细节仍是锤炼阶段重点风险。

## 公开评估

- evidence_public: false
- 不公开原因：当前只有一个小型推演脚本，主要服务正文表格生成；没有独立成体系的实验项目。后续初稿若直接引用脚本输出数字，可在 Step 2.5 重新评估是否公开。

## 正文使用提醒

1. `evidence/data/local-tcp-keepalive-snapshot.md` 是 macOS 实测，不要写成 Linux 实测。
2. `evidence/output/keepalive-matrix/result.md` 是推演表，正文可自然描述为“我把几组配置放在同一张表里算了一下”，不要伪装成生产压测。
3. “5 分钟”除非补充现场证据，否则正文统一写成“5 分钟级”或“分钟级”。
4. E1 不足时，用 E3/E5 承担核心证据，不编造真实生产细节。
