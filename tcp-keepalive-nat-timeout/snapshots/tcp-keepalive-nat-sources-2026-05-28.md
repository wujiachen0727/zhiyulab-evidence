# TCP keepalive / NAT idle timeout 求证快照

验证日期：2026-05-28

## 1. Linux TCP keepalive 默认参数

来源：Linux tcp(7) manual page
URL：https://man7.org/linux/man-pages/man7/tcp.7.html

可引用事实：
- `tcp_keepalive_time`：启用 `SO_KEEPALIVE` 后，连接空闲多少秒后开始发送 keepalive probe；默认 7200 秒（2 小时）。
- `tcp_keepalive_intvl`：keepalive probe 之间的间隔；默认 75 秒。
- `tcp_keepalive_probes`：放弃并杀死连接前发送的 probe 数；默认 9。
- 单个 socket 可通过 `TCP_KEEPIDLE`、`TCP_KEEPINTVL`、`TCP_KEEPCNT` 覆盖系统默认值。

## 2. AWS NAT Gateway TCP idle timeout

来源：AWS VPC NAT Gateway troubleshooting
URL：https://docs.aws.amazon.com/vpc/latest/userguide/nat-gateway-troubleshooting.html

可引用事实：
- NAT Gateway 在连接空闲 350 秒或更长时间后可能导致连接超时。
- 超时后，NAT Gateway 会向继续使用该连接的资源返回 RST 数据包，而不是 FIN。
- AWS 建议在实例上启用 TCP keepalive，并将 keepalive 设置为小于 350 秒。

## 3. Azure Load Balancer TCP idle timeout

来源：Microsoft Learn — Configure load balancer TCP reset and idle timeout
URL：https://learn.microsoft.com/en-us/azure/load-balancer/load-balancer-tcp-idle-timeout

可引用事实：
- Azure Load Balancer TCP idle timeout 默认 4 分钟，可配置范围 4 到 100 分钟。
- 不活动时间超过 timeout 后，不保证 TCP/HTTP 会话继续保持。
- 可启用 TCP Reset，让连接关闭行为更显式。

## 4. 当前文章写作边界

“TCP keepalive 2h + NAT 5min = 连接被静默杀死”适合作为故障公式，但正文需要严谨表述：

- 2h：Linux keepalive 默认值，有官方手册支撑。
- 5min：应写成“5 分钟级 NAT/LB 空闲回收”，因为不同基础设施默认值不同：Azure 为 4 分钟，AWS NAT Gateway 为 350 秒。
- 如果作者实际环境就是 5 分钟，需要在论证阶段补充实测或配置截图作为自造证据。
