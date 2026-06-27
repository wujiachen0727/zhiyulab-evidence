# E1: ZSET 编码切换实测

## 运行环境

- Redis 8.8.0 (homebrew)
- macOS Darwin 25.5.0 ARM64

## 运行步骤

```bash
# 启动测试 Redis（端口 16379，避免冲突）
redis-server --port 16379 --daemonize yes --save "" --appendonly no

# 运行测试
bash test.sh

# 清理
redis-cli -p 16379 SHUTDOWN NOSAVE
```

## 预期输出

见 `../../output/zset-encoding-test-output.txt`

关键数据：
- 128 个元素时编码为 listpack，内存 1596 字节
- 129 个元素时编码切换为 skiplist，内存 12576 字节（7.9x 跃升）
- 单元素大小超过 64 字节时触发切换
- 配置项：zset-max-listpack-entries=128, zset-max-listpack-value=64
