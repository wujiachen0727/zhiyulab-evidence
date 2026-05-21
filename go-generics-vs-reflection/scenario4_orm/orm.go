// 场景4：ORM/数据库映射 —— 反射版 vs 泛型版
package scenario4

import (
	"fmt"
	"reflect"
	"strings"
)

// 测试用结构体
type Article struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	Content string `db:"content"`
	Author  string `db:"author"`
}

// === 反射版 ORM：查询构建 + 结果映射都用反射 ===

func SelectReflect(table string, dest interface{}, where string) string {
	// 通过反射读取 struct tag 构建 SELECT 语句
	t := reflect.TypeOf(dest)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag != "" {
			columns = append(columns, tag)
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	if where != "" {
		sql += " WHERE " + where
	}
	return sql
}

// 模拟结果映射（反射版 — 运行时动态赋值）
func ScanRowReflect(dest interface{}, columns []string, values []interface{}) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	// 构建 tag -> field index 映射
	tagMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag != "" {
			tagMap[tag] = i
		}
	}

	// 动态赋值
	for i, col := range columns {
		if idx, ok := tagMap[col]; ok {
			field := v.Field(idx)
			val := reflect.ValueOf(values[i])
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
			}
		}
	}
	return nil
}

// === 泛型版：查询构建类型安全，结果映射仍需反射 ===

type Table[T any] struct {
	name string
}

func NewTable[T any](name string) *Table[T] {
	return &Table[T]{name: name}
}

// 泛型查询构建器：编译时确定返回类型
func (t *Table[T]) Select(where string) *Query[T] {
	return &Query[T]{table: t.name, where: where}
}

type Query[T any] struct {
	table string
	where string
}

// 构建 SQL（仍需反射读 tag — 类型信息不含 tag 内容）
func (q *Query[T]) BuildSQL() string {
	var zero T
	t := reflect.TypeOf(zero)

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag != "" {
			columns = append(columns, tag)
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), q.table)
	if q.where != "" {
		sql += " WHERE " + q.where
	}
	return sql
}

// 结果映射（泛型确保返回正确类型，内部仍用反射）
func (q *Query[T]) ScanRow(columns []string, values []interface{}) (T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()
	t := v.Type()

	tagMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag != "" {
			tagMap[tag] = i
		}
	}

	for i, col := range columns {
		if idx, ok := tagMap[col]; ok {
			field := v.Field(idx)
			val := reflect.ValueOf(values[i])
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
			}
		}
	}
	return result, nil // 返回类型是编译时确定的 T，不需要类型断言
}
