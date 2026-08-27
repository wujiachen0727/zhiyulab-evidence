package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func BenchmarkUnmarshalLarge(b *testing.B) {
	payload := []byte(`{"id":1,"name":"bench","tags":["a","b","c"],"meta":{"k":"v"}}`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var out map[string]any
		if err := json.Unmarshal(payload, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func TestDuplicateKey(t *testing.T) {
	var out map[string]any
	err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["a"] != float64(2) {
		t.Fatalf("last key wins: got %v", out["a"])
	}
}

func TestInvalidUTF8(t *testing.T) {
	var out map[string]any
	err := json.Unmarshal([]byte(`{"x":"\xff\xfe"}`), &out)
	if err == nil {
		t.Fatal("expected error for invalid utf-8 in string")
	}
}

func TestTypeMismatchError(t *testing.T) {
	var out struct {
		N int `json:"n"`
	}
	err := json.Unmarshal([]byte(`{"n":"not-a-number"}`), &out)
	if err == nil {
		t.Fatal("expected type mismatch error")
	}
	if !strings.Contains(err.Error(), "json") {
		t.Fatalf("unexpected error text: %q", err.Error())
	}
}
