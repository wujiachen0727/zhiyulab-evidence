# E5: Java JPMS 三特征框架检验

> 类型：场景模拟 | 用判别框架检验"好意图坏执行"

## 场景设定

一个中型 Java 项目（30+ 模块）从 Java 8 迁移到 Java 17，需要引入 JPMS。

## 特征 1 检验：错误路径变丑？

### "错误路径"：不使用模块系统

```java
// Java 8 风格：任何包都可以访问任何公开类
// 想用 internal API？直接 import：
import com.sun.internal.SomeUtil; // 只是 deprecation 警告
```

### "被摩擦后的路径"：使用模块系统

```java
// module-info.java
module my.app {
    requires java.base;
    requires java.sql;
    requires com.fasterxml.jackson.databind;
    exports my.app.api;
    opens my.app.model to com.fasterxml.jackson.databind;
}
```

**判定：✅ 通过（2/2）**

错误路径（随意依赖 internal API）确实"变丑"了——编译器直接拒绝。
但问题在于下一条特征...

## 特征 2 检验：正确路径保持简洁？

### "正确做法"的实际体验

想要正确地使用模块系统，你需要：

1. **每个模块写 module-info.java**（即使只有 3 个类的小模块）
2. **处理 split package 问题**：两个 JAR 不能有同名包
3. **处理反射访问**：Spring/Hibernate 需要 `opens ... to ...`
4. **处理自动模块**：没有 module-info 的第三方 JAR 变成"自动模块"
5. **配置 --add-opens / --add-reads**：框架内部需要

### 对比：不用 JPMS 的体验

```java
// 什么都不写，直接用 classpath。照样运行。
// 大部分项目选择了这条路。
```

**判定：❌ 不通过（0/2）**

"正确路径"（用 JPMS）比"旧路径"（classpath）复杂 5-10 倍。
**体验落差方向反了**——新的"正确做法"比旧做法更痛苦。

## 特征 3 检验：使用者认知提升？

### 写了 module-info 后你理解了什么？

最佳情况：
- ✅ 理解了模块之间的依赖关系应该是显式的
- ✅ 理解了"不是所有公开类都应该对外暴露"

最常见情况：
- ❌ "我加了这行是因为 Spring 启动要 --add-opens"
- ❌ "我不知道为什么需要 opens to，但没它报错"
- ❌ "我把所有包都 exports 了，因为不知道哪些需要"

**判定：⚠️ 部分通过（1/2）**

设计意图上有认知提升价值，但实际执行中大部分开发者在"应付编译器"而非"学习设计"。

## 总分

| 特征 | 得分 | 说明 |
|------|:----:|------|
| 1. 错误路径变丑 | 2 | 直接编译拒绝，确实很丑 |
| 2. 正确路径简洁 | 0 | 正确路径反而更复杂 |
| 3. 认知提升 | 1 | 有潜力但实际体验多是"配置驱动" |
| **总分** | **3** | 边界线——好意图，坏执行 |

## 为什么 JPMS 失败了？

关键区别：Go reflect 和 Rust unsafe 的"正确路径"是**语言本身的正常用法**——你不用 reflect 时代码天然简洁，不用 unsafe 时编译器帮你做了安全检查。

而 Java JPMS 的"正确路径"需要**额外的配置成本**——你需要写 module-info、处理 split package、配置 --add-opens。这不是"走正确路径的奖励"，而是"使用新特性的入场券"。

**一句话总结**：好的摩擦力让正确路径免费，坏的摩擦力让所有路径都收费。
