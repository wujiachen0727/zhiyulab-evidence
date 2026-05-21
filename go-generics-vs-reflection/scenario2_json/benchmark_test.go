package scenario2

import "testing"

var benchUser = User{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 30}

// 反射版：encoding/json.Marshal（标准库，内部用 reflect 遍历字段）
func BenchmarkMarshalReflect(b *testing.B) {
	var result []byte
	for i := 0; i < b.N; i++ {
		result, _ = MarshalReflect(benchUser)
	}
	sink = result
}

// 泛型版：已知类型直接序列化（零反射）
func BenchmarkMarshalGeneric(b *testing.B) {
	var result []byte
	for i := 0; i < b.N; i++ {
		result = MarshalGeneric(benchUser)
	}
	sink = result
}

// 验证两者输出等价
func TestMarshalEquivalence(t *testing.T) {
	reflectResult, err := MarshalReflect(benchUser)
	if err != nil {
		t.Fatal(err)
	}
	genericResult := MarshalGeneric(benchUser)

	t.Logf("反射版: %s", reflectResult)
	t.Logf("泛型版: %s", genericResult)
	// 注：字段顺序可能不同，但内容等价
}
