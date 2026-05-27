# E2: Rust safe vs unsafe 编译器行为对比

> 降级：环境限制（rustc 未安装），改为伪代码+文档推演

## 实验设计

对比同一操作在 safe 和 unsafe 语境下编译器的不同反应。

## 场景 1：解引用裸指针

### Safe 版（编译器拒绝）

```rust
fn main() {
    let x: i32 = 42;
    let ptr: *const i32 = &x;
    let val = *ptr;  // ERROR: 编译器直接拒绝
    println!("{}", val);
}
```

**编译器输出**（来自 Rust Reference）：
```
error[E0133]: dereference of raw pointer is unsafe and requires unsafe function or block
 --> src/main.rs:4:15
  |
4 |     let val = *ptr;
  |               ^^^^ dereference of raw pointer
  |
  = note: raw pointers may be null, dangling or unaligned;
          they can violate aliasing rules and cause data races
```

### Unsafe 版（编译器允许，但需要签名）

```rust
fn main() {
    let x: i32 = 42;
    let ptr: *const i32 = &x;
    // SAFETY: ptr was derived from a valid reference to x,
    // which is still in scope and properly aligned.
    let val = unsafe { *ptr };
    println!("{}", val);
}
```

**关键设计**：
1. `unsafe` 块 = "我知道我在做什么"的声明
2. `// SAFETY:` 注释 = 社区规范（clippy lint 会警告缺失）
3. 范围最小化 = 只包裹需要 unsafe 的那一行

## 场景 2：调用 unsafe 函数

### 未用 unsafe 包裹

```rust
use std::slice;

fn main() {
    let ptr = 0x1234 as *const u8;
    let s = slice::from_raw_parts(ptr, 10);  // ERROR
}
```

**编译器输出**：
```
error[E0133]: call to unsafe function `std::slice::from_raw_parts` is unsafe
 --> src/main.rs:5:13
  |
5 |     let s = slice::from_raw_parts(ptr, 10);
  |             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ call to unsafe function
```

### 正确用法

```rust
use std::slice;

fn main() {
    let data: [u8; 10] = [0; 10];
    let ptr = data.as_ptr();
    // SAFETY: ptr points to a valid [u8; 10] array still in scope,
    // and we're requesting exactly 10 bytes (matching the source).
    let s = unsafe { slice::from_raw_parts(ptr, 10) };
    assert_eq!(s.len(), 10);
}
```

## unsafe 五种超能力（来自 Rust Nomicon）

在 unsafe 块中可以做、但 safe Rust 中不允许的 5 件事：

1. **解引用裸指针**（`*const T` / `*mut T`）
2. **调用 unsafe 函数**（包括 FFI 和带 `unsafe` 标记的函数）
3. **访问或修改可变静态变量**
4. **实现 unsafe trait**
5. **访问 union 的字段**

## 摩擦力分析

| 维度 | 表现 |
|------|------|
| 语法信号 | `unsafe {}` 关键字块——像红色警告牌 |
| 社区规范 | `// SAFETY:` 注释——解释为什么这段代码是安全的 |
| 范围限制 | unsafe 块应尽可能小——只包裹必要代码 |
| 工具支持 | clippy::undocumented_unsafe_blocks lint |
| 审查焦点 | Code review 优先检查 unsafe 块 |

## 与 Go reflect 摩擦力的对比

| 维度 | Go reflect | Rust unsafe |
|------|-----------|-------------|
| 摩擦方式 | API 冗长（运行时检查） | 语法标记（编译时阻断） |
| 阻断时机 | 运行时 panic | 编译时拒绝 |
| 信号强度 | 中（代码变长变丑） | 高（红色关键字+强制注释） |
| "签名"方式 | 无（你自己负责） | 显式（`// SAFETY:` 注释） |
| 失败代价 | panic + 运行时类型错误 | 未定义行为（更严重） |

## 结论

Rust unsafe 是比 Go reflect 更"重"的摩擦力设计：
- Go 让你"感觉到贵"（认知税）
- Rust 让你"为每一行签名担责"（安全仪式）

两者都是有意设计，但"重量级"不同。unsafe 的代价更高（UB），所以摩擦力也更重。
