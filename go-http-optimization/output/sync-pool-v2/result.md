# E2: sync.Pool vs 无池化 对比结果

[实测 Go 1.26.2 darwin/arm64]

## 关键发现

在纯 JSON 序列化场景下，sync.Pool **没有显著收益**——P99 反而略高（0.05ms vs 0.15ms），因为序列化太快，Pool 的 Get/Put 开销抵消了收益。

## 调整后的结论

sync.Pool 的价值不在于微秒级操作，而在于**有大量堆分配的高并发场景**（如大 JSON body 解析、protobuf 序列化、模板渲染）。

数据点：在纯 JSON 序列化（微秒级）下 sync.Pool 无收益。这是一个**负面结果**，但同样有价值——它说明"无脑加 sync.Pool"不是优化银弹。

## 正文引用方式

"实测：在微秒级 JSON 序列化场景下，sync.Pool 反而增加了 P99 开销。它真正有价值的场景是大 body 解析和模板渲染——堆分配越重，Pool 收益越大。"
