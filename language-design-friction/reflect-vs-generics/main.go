// Package main demonstrates the friction difference between
// Go reflect API and generics for the same task:
// "get a struct field by name and return its string representation"
//
// Purpose: Show that reflect requires significantly more boilerplate
// and ceremony, which is a deliberate design choice to signal cost.
package main

import (
	"fmt"
	"reflect"
)

// --- Domain types ---

type User struct {
	Name  string
	Email string
	Age   int
}

// --- Approach 1: Reflect (verbose, unsafe at runtime) ---

// GetFieldReflect retrieves a struct field value by name using reflect.
// Note the ceremony: 3 layers of safety checks before you can read a value.
func GetFieldReflect(obj any, fieldName string) (string, error) {
	v := reflect.ValueOf(obj)

	// Layer 1: Must be a struct (or pointer to struct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected struct, got %s", v.Kind())
	}

	// Layer 2: Field must exist
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return "", fmt.Errorf("no such field: %s", fieldName)
	}

	// Layer 3: Field must be exportable
	if !field.CanInterface() {
		return "", fmt.Errorf("field %s is unexported", fieldName)
	}

	return fmt.Sprintf("%v", field.Interface()), nil
}

// --- Approach 2: Generics (concise, type-safe at compile time) ---

// FieldGetter is a constraint for types that can describe their fields.
type FieldGetter interface {
	GetField(name string) string
}

// User satisfies FieldGetter
func (u User) GetField(name string) string {
	switch name {
	case "Name":
		return u.Name
	case "Email":
		return u.Email
	case "Age":
		return fmt.Sprintf("%d", u.Age)
	default:
		return ""
	}
}

// GetFieldGeneric uses a compile-time constraint.
// No runtime type checks. No reflection. Zero ceremony.
func GetFieldGeneric[T FieldGetter](obj T, fieldName string) string {
	return obj.GetField(fieldName)
}

func main() {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}

	// Reflect approach
	val, err := GetFieldReflect(u, "Name")
	fmt.Printf("[reflect] Name = %s (err: %v)\n", val, err)

	// Generics approach
	val2 := GetFieldGeneric(u, "Name")
	fmt.Printf("[generics] Name = %s\n", val2)
}
