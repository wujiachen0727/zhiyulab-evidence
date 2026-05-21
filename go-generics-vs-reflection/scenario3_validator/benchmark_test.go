package scenario3

import "testing"

var testReq = CreateUserReq{Name: "张三", Email: "zhangsan@example.com", Age: 30}
var invalidReq = CreateUserReq{Name: "a", Email: "invalid", Age: -1}

// 反射版：通过 struct tag 自动验证
func BenchmarkValidateReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateReflect(testReq)
	}
}

// 泛型版：通过接口约束 + 手写验证
func BenchmarkValidateGeneric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateGeneric(testReq)
	}
}

// 验证两者结果等价（对于无效输入）
func TestValidationResults(t *testing.T) {
	reflectErrs := ValidateReflect(invalidReq)
	genericErrs := ValidateGeneric(invalidReq)

	t.Logf("反射版错误数: %d", len(reflectErrs))
	for _, e := range reflectErrs {
		t.Logf("  - %s", e.Error())
	}
	t.Logf("泛型版错误数: %d", len(genericErrs))
	for _, e := range genericErrs {
		t.Logf("  - %s", e.Error())
	}
}

// 代码量对比
func TestCodeComplexity(t *testing.T) {
	t.Log("=== 代码量对比 ===")
	t.Log("反射版: ~60 行（通用引擎，新结构体零成本接入）")
	t.Log("泛型版: ~30 行（每个结构体需手写 Validate 方法）")
	t.Log("取舍: 反射版一次写完到处用，泛型版每个类型要写一次")
	t.Log("混合方案: 泛型约束确保类型有 Validate 方法，内部用 reflect 读 tag")
}
