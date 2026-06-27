# 环境预检记录

**日期**：2026-05-29

| 环境 | 结果 | 说明 |
|------|------|------|
| Go | ✅ go1.26.2 darwin/arm64 | 已安装，直接运行 Go 实验 |
| Java | ✅ OpenJDK 26.0.1 | 初始不可用，已通过 Homebrew 安装 `openjdk`；运行时使用 `/opt/homebrew/opt/openjdk/bin/java` |
| Erlang | ✅ Erlang/OTP 29 | 初始不可用，已通过 Homebrew 安装 `erlang`；使用 `escript` 运行实验 |

## 影响

E1/E2 不需要降级为伪实现。Go、Java virtual thread、Erlang/Actor 风格均有可运行代码和输出。
