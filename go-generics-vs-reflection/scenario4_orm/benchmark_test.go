package scenario4

import "testing"

// 模拟数据库返回的一行数据
var mockColumns = []string{"id", "title", "content", "author"}
var mockValues = []interface{}{1, "Go泛型实战", "本文探讨...", "张三"}

// 反射版：查询构建
func BenchmarkSelectReflect(b *testing.B) {
	var articles []Article
	for i := 0; i < b.N; i++ {
		_ = SelectReflect("articles", &articles, "author = '张三'")
	}
}

// 泛型版：查询构建
func BenchmarkSelectGeneric(b *testing.B) {
	table := NewTable[Article]("articles")
	for i := 0; i < b.N; i++ {
		q := table.Select("author = '张三'")
		_ = q.BuildSQL()
	}
}

// 反射版：结果映射
func BenchmarkScanReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var a Article
		_ = ScanRowReflect(&a, mockColumns, mockValues)
	}
}

// 泛型版：结果映射（内部仍反射，但返回类型安全）
func BenchmarkScanGeneric(b *testing.B) {
	table := NewTable[Article]("articles")
	q := table.Select("")
	for i := 0; i < b.N; i++ {
		_, _ = q.ScanRow(mockColumns, mockValues)
	}
}

func TestORMComparison(t *testing.T) {
	// 反射版
	var articles []Article
	sql1 := SelectReflect("articles", &articles, "author = '张三'")
	t.Logf("反射版 SQL: %s", sql1)

	var a1 Article
	_ = ScanRowReflect(&a1, mockColumns, mockValues)
	t.Logf("反射版结果: %+v", a1)

	// 泛型版
	table := NewTable[Article]("articles")
	q := table.Select("author = '张三'")
	sql2 := q.BuildSQL()
	t.Logf("泛型版 SQL: %s", sql2)

	a2, _ := q.ScanRow(mockColumns, mockValues)
	t.Logf("泛型版结果: %+v", a2)

	// 关键差异
	t.Log("\n=== 关键洞察 ===")
	t.Log("1. 查询构建: 泛型版返回 *Query[Article]，编译时确定结果类型")
	t.Log("2. 结果映射: 两版都需要反射（struct tag 是运行时元数据）")
	t.Log("3. 泛型版优势: 调用方不需要 .(*Article) 类型断言")
}
