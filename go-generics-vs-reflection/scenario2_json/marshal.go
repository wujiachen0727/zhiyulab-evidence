// 场景2：JSON/序列化 —— 反射版 vs 泛型版
// 对比：已知类型场景下，泛型直接编码 vs encoding/json 反射编码
package scenario2

import (
	"encoding/json"
	"strconv"
	"unsafe"
)

// 测试用结构体
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// === 反射版：标准库 encoding/json（内部使用 reflect）===
func MarshalReflect(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// === 泛型版：已知类型的手写序列化（避免反射）===
// 这是泛型能做到的极限：通过类型约束接口实现零反射序列化

type JSONMarshaler interface {
	MarshalFast() []byte
}

func MarshalGeneric[T JSONMarshaler](v T) []byte {
	return v.MarshalFast()
}

// User 实现 JSONMarshaler（手写，零反射）
func (u User) MarshalFast() []byte {
	// 手动拼接 JSON — 完全避免反射
	buf := make([]byte, 0, 128)
	buf = append(buf, '{')
	buf = append(buf, `"id":`...)
	buf = strconv.AppendInt(buf, int64(u.ID), 10)
	buf = append(buf, `,"name":"`...)
	buf = appendEscaped(buf, u.Name)
	buf = append(buf, `","email":"`...)
	buf = appendEscaped(buf, u.Email)
	buf = append(buf, `","age":`...)
	buf = strconv.AppendInt(buf, int64(u.Age), 10)
	buf = append(buf, '}')
	return buf
}

// === 未知类型场景：仍需反射 ===
// 当输入是 interface{} 时，泛型无法帮忙——因为编译时不知道字段结构
func MarshalUnknown(v interface{}) ([]byte, error) {
	// 只能回到标准库
	return json.Marshal(v)
}

// 辅助函数：简单的 JSON 字符串转义
func appendEscaped(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '\\':
			buf = append(buf, '\\', c)
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			buf = append(buf, c)
		}
	}
	return buf
}

// 用于防止编译器优化掉结果
var sink []byte

func init() {
	_ = unsafe.Sizeof(sink)
}
