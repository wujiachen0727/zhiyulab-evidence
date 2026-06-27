#!/usr/bin/env python3
"""生成 TCP keepalive 与中间设备 idle timeout 的配置对比表。"""

from dataclasses import dataclass
from typing import Optional


@dataclass(frozen=True)
class Scenario:
    name: str
    middle_idle_s: int
    keepalive_idle_s: Optional[int]
    keepalive_interval_s: Optional[int]
    keepalive_probes: Optional[int]
    pool_idle_s: Optional[int]
    note: str

    def first_probe_s(self) -> str:
        if self.keepalive_idle_s is None:
            return "未启用"
        return f"{self.keepalive_idle_s}s"

    def tcp_dead_s(self) -> str:
        if self.keepalive_idle_s is None or self.keepalive_interval_s is None or self.keepalive_probes is None:
            return "不会由 TCP keepalive 主动发现"
        return f"{self.keepalive_idle_s + self.keepalive_interval_s * self.keepalive_probes}s"

    def pool_retires_before_middle(self) -> bool:
        return self.pool_idle_s is not None and self.pool_idle_s < self.middle_idle_s

    def probe_before_middle(self) -> bool:
        return self.keepalive_idle_s is not None and self.keepalive_idle_s < self.middle_idle_s

    def verdict(self) -> str:
        if self.pool_retires_before_middle():
            return "连接池先回收：安全边界更靠前"
        if self.probe_before_middle():
            return "TCP 探测早于中间设备回收"
        return "探测晚于回收：存在陈旧连接窗口"


SCENARIOS = [
    Scenario(
        name="Linux 默认 keepalive + 5 分钟级 NAT",
        middle_idle_s=300,
        keepalive_idle_s=7200,
        keepalive_interval_s=75,
        keepalive_probes=9,
        pool_idle_s=None,
        note="Linux tcp(7) 默认参数；SO_KEEPALIVE 开启后仍要等 2 小时才首次探测",
    ),
    Scenario(
        name="AWS NAT Gateway 350s + Linux 默认 keepalive",
        middle_idle_s=350,
        keepalive_idle_s=7200,
        keepalive_interval_s=75,
        keepalive_probes=9,
        pool_idle_s=None,
        note="AWS 文档示例；350s 仍远早于 7200s",
    ),
    Scenario(
        name="Azure LB 默认 240s + Linux 默认 keepalive",
        middle_idle_s=240,
        keepalive_idle_s=7200,
        keepalive_interval_s=75,
        keepalive_probes=9,
        pool_idle_s=None,
        note="Azure 默认 4 分钟；不能把 5 分钟写成通用默认值",
    ),
    Scenario(
        name="调低 TCP keepalive：60/10/3",
        middle_idle_s=300,
        keepalive_idle_s=60,
        keepalive_interval_s=10,
        keepalive_probes=3,
        pool_idle_s=None,
        note="首次探测早于 300s，最晚约 90s 判死；适合需要长连接保活的场景",
    ),
    Scenario(
        name="连接池 idle timeout 240s",
        middle_idle_s=300,
        keepalive_idle_s=7200,
        keepalive_interval_s=75,
        keepalive_probes=9,
        pool_idle_s=240,
        note="不依赖 TCP 探活，应用侧先丢弃旧连接",
    ),
    Scenario(
        name="错误调参：keepalive_idle=600s",
        middle_idle_s=300,
        keepalive_idle_s=600,
        keepalive_interval_s=30,
        keepalive_probes=3,
        pool_idle_s=None,
        note="看似调低了 2 小时，但仍晚于 5 分钟级回收",
    ),
]


def main() -> None:
    print("# keepalive / idle timeout 配置组合对比")
    print()
    print("标注：本表为 [推演]，用于解释不同配置组合的时间关系；Linux 默认参数来自 tcp(7) 文档事实，云厂商 idle timeout 来自立意阶段求证快照。")
    print()
    print("| 场景 | 中间设备 idle timeout | TCP 首次 keepalive | TCP 最晚判死时间 | 连接池 idle timeout | 判定 | 说明 |")
    print("|---|---:|---:|---:|---:|---|---|")
    for scenario in SCENARIOS:
        pool = "未设置" if scenario.pool_idle_s is None else f"{scenario.pool_idle_s}s"
        print(
            f"| {scenario.name} | {scenario.middle_idle_s}s | {scenario.first_probe_s()} | "
            f"{scenario.tcp_dead_s()} | {pool} | {scenario.verdict()} | {scenario.note} |"
        )
    print()
    print("## 可供正文引用的结论")
    print()
    print("1. 如果中间设备 300s 左右回收空闲连接，而 TCP keepalive 7200s 才首次探测，那么 TCP 探测一定来得太晚。")
    print("2. 修复原则不是固定改成某个数字，而是让主动探测或连接池回收发生在中间设备回收之前。")
    print("3. 只把 keepalive 从 7200s 改成 600s 仍可能无效，因为它依然晚于 300s 的回收窗口。")


if __name__ == "__main__":
    main()
