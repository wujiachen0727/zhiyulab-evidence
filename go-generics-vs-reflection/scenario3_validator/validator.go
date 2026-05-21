// 场景3：结构体验证 —— 反射版 validator vs 泛型约束版
package scenario3

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// 测试用结构体
type CreateUserReq struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"required,min=0,max=150"`
}

// === 反射版：读取 struct tag 执行验证 ===

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateReflect(v interface{}) []ValidationError {
	var errs []ValidationError
	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}
		fieldVal := val.Field(i)
		rules := strings.Split(tag, ",")

		for _, rule := range rules {
			if err := applyRule(field.Name, fieldVal, rule); err != nil {
				errs = append(errs, *err)
			}
		}
	}
	return errs
}

func applyRule(name string, val reflect.Value, rule string) *ValidationError {
	switch {
	case rule == "required":
		if val.IsZero() {
			return &ValidationError{name, "is required"}
		}
	case strings.HasPrefix(rule, "min="):
		min, _ := strconv.Atoi(rule[4:])
		switch val.Kind() {
		case reflect.String:
			if val.Len() < min {
				return &ValidationError{name, fmt.Sprintf("min length %d", min)}
			}
		case reflect.Int, reflect.Int64:
			if val.Int() < int64(min) {
				return &ValidationError{name, fmt.Sprintf("min value %d", min)}
			}
		}
	case strings.HasPrefix(rule, "max="):
		max, _ := strconv.Atoi(rule[4:])
		switch val.Kind() {
		case reflect.String:
			if val.Len() > max {
				return &ValidationError{name, fmt.Sprintf("max length %d", max)}
			}
		case reflect.Int, reflect.Int64:
			if val.Int() > int64(max) {
				return &ValidationError{name, fmt.Sprintf("max value %d", max)}
			}
		}
	case rule == "email":
		if val.Kind() == reflect.String && !strings.Contains(val.String(), "@") {
			return &ValidationError{name, "invalid email"}
		}
	}
	return nil
}

// === 泛型版：类型约束 + 编译时检查 ===
// 泛型能做的：通过接口约束确保类型实现了 Validate() 方法
// 泛型做不到的：自动读取 struct tag（这是元数据，编译时不可遍历）

type Validatable interface {
	Validate() []ValidationError
}

func ValidateGeneric[T Validatable](v T) []ValidationError {
	return v.Validate()
}

// CreateUserReq 手动实现验证（无反射，但需要为每个结构体手写）
func (r CreateUserReq) Validate() []ValidationError {
	var errs []ValidationError
	if r.Name == "" {
		errs = append(errs, ValidationError{"Name", "is required"})
	} else if len(r.Name) < 2 {
		errs = append(errs, ValidationError{"Name", "min length 2"})
	} else if len(r.Name) > 50 {
		errs = append(errs, ValidationError{"Name", "max length 50"})
	}
	if r.Email == "" {
		errs = append(errs, ValidationError{"Email", "is required"})
	} else if !strings.Contains(r.Email, "@") {
		errs = append(errs, ValidationError{"Email", "invalid email"})
	}
	if r.Age < 0 {
		errs = append(errs, ValidationError{"Age", "min value 0"})
	} else if r.Age > 150 {
		errs = append(errs, ValidationError{"Age", "max value 150"})
	}
	return errs
}
