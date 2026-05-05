// 泛型 vs 代码生成对比实验
// 场景：实现一个类型安全的 Set 数据结构
// 对比两种实现方式的代码量和可维护性

package main

// ============================================================
// 方案 A：使用泛型（Go 1.18+）
// ============================================================

// Set 是一个泛型集合，支持任何 comparable 类型
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{m: make(map[T]struct{})}
}

func (s *Set[T]) Add(v T)          { s.m[v] = struct{}{} }
func (s *Set[T]) Remove(v T)       { delete(s.m, v) }
func (s *Set[T]) Contains(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int         { return len(s.m) }

// 泛型方案总计：~15 行有效代码
// 支持所有 comparable 类型，无需为每个类型生成代码
// 编译时类型检查 ✅
// 可读性 ✅
// 维护成本：改一处，所有类型自动更新

// ============================================================
// 方案 B：使用代码生成（text/template 生成特定类型的 Set）
// ============================================================

// 需要：
// 1. 一个 set.tmpl 模板文件（~30 行）
// 2. 一个 gen.go 生成器脚本（~50 行）
// 3. 每个类型生成一个 xxx_set.go 文件
// 4. 一个 //go:generate 指令

// 生成的代码示例（int_set.go）：
/*
type IntSet struct {
    m map[int]struct{}
}

func NewIntSet() *IntSet {
    return &IntSet{m: make(map[int]struct{})}
}

func (s *IntSet) Add(v int)          { s.m[v] = struct{}{} }
func (s *IntSet) Remove(v int)       { delete(s.m, v) }
func (s *IntSet) Contains(v int) bool { _, ok := s.m[v]; return ok }
func (s *IntSet) Len() int            { return len(s.m) }
*/

// 代码生成方案总计：
// - 模板文件：~30 行
// - 生成器脚本：~50 行
// - 每新增一种类型：+1 行 generate 指令 + 生成 ~15 行代码
// - 修改接口：需要改模板 + 重新生成所有文件
// - 维护成本：改模板 → 重新 generate → 检查所有生成文件

// ============================================================
// 对比结论
// ============================================================
//
// | 维度          | 泛型方案    | 代码生成方案      |
// |---------------|------------|------------------|
// | 初始代码量     | ~15 行     | ~80 行（模板+生成器）|
// | 新增类型成本   | 0 行       | +1 行指令         |
// | 修改接口成本   | 改 1 处    | 改模板+重新生成    |
// | 编译时类型安全  | ✅         | ✅               |
// | CI 额外步骤   | 无         | 需要 go generate  |
// | 可读性        | 高         | 中（需看模板+生成文件）|
//
// 判定：当需求是"同一接口用于多种类型"时，泛型完胜。
// 代码生成在这类场景是 overengineering。
//
// 什么时候 codegen 仍然必要？
// - 每种类型需要生成【不同结构】的代码（不只是类型参数不同）
// - 需要基于类型的字段列表生成代码（如 ORM、序列化）
// - 需要为外部 schema 生成 Go 类型定义
