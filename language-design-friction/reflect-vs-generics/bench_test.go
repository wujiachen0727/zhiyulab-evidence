package main

import (
	"fmt"
	"reflect"
	"testing"
)

type BenchUser struct {
	Name  string
	Email string
	Age   int
	Role  string
	Team  string
}

func (u BenchUser) GetField(name string) string {
	switch name {
	case "Name":
		return u.Name
	case "Email":
		return u.Email
	case "Age":
		return fmt.Sprintf("%d", u.Age)
	case "Role":
		return u.Role
	case "Team":
		return u.Team
	default:
		return ""
	}
}

var benchUser = BenchUser{
	Name:  "Alice",
	Email: "alice@example.com",
	Age:   30,
	Role:  "Engineer",
	Team:  "Platform",
}

// Benchmark: reflect approach
func BenchmarkGetFieldReflect(b *testing.B) {
	for b.Loop() {
		GetFieldReflect(benchUser, "Name")
	}
}

// Benchmark: generics approach
func BenchmarkGetFieldGeneric(b *testing.B) {
	for b.Loop() {
		GetFieldGeneric(benchUser, "Name")
	}
}

// Benchmark: reflect with type assertion (common pattern)
func BenchmarkReflectTypeCheck(b *testing.B) {
	for b.Loop() {
		v := reflect.ValueOf(benchUser)
		_ = v.Kind() == reflect.Struct
	}
}

// Benchmark: reflect full traversal (iterate all fields)
func BenchmarkReflectTraverseAll(b *testing.B) {
	for b.Loop() {
		v := reflect.ValueOf(benchUser)
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			_ = v.Field(i).Interface()
		}
	}
}

// Benchmark: direct field access (baseline)
func BenchmarkDirectAccess(b *testing.B) {
	for b.Loop() {
		_ = benchUser.Name
	}
}
