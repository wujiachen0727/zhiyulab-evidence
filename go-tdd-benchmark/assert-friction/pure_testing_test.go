package assertfriction

import "testing"

// 纯 testing 包写法：5 个断言场景
func TestAdd_PureTesting(t *testing.T) {
	got := Add(1, 2)
	if got != 3 {
		t.Errorf("Add(1, 2) = %d, want 3", got)
	}
}

func TestGreeting_PureTesting(t *testing.T) {
	got := Greeting("Go")
	if got != "Hello, Go" {
		t.Errorf("Greeting(%q) = %q, want %q", "Go", got, "Hello, Go")
	}
}

func TestDivide_PureTesting(t *testing.T) {
	got, err := Divide(10, 3)
	if err != nil {
		t.Fatalf("Divide(10, 3) returned error: %v", err)
	}
	if got != 3.3333333333333335 {
		t.Errorf("Divide(10, 3) = %f, want 3.333333", got)
	}
}

func TestDivideByZero_PureTesting(t *testing.T) {
	_, err := Divide(10, 0)
	if err == nil {
		t.Fatal("Divide(10, 0) should return error, got nil")
	}
}

func TestItems_PureTesting(t *testing.T) {
	got := Items()
	if len(got) != 3 {
		t.Fatalf("Items() returned %d items, want 3", len(got))
	}
	if got[0] != "a" {
		t.Errorf("Items()[0] = %q, want %q", got[0], "a")
	}
	if got[1] != "b" {
		t.Errorf("Items()[1] = %q, want %q", got[1], "b")
	}
	if got[2] != "c" {
		t.Errorf("Items()[2] = %q, want %q", got[2], "c")
	}
}
