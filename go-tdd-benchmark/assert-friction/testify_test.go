package assertfriction

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// testify 写法：5 个断言场景
func TestAdd_Testify(t *testing.T) {
	assert.Equal(t, 3, Add(1, 2), "Add(1, 2) should equal 3")
}

func TestGreeting_Testify(t *testing.T) {
	assert.Equal(t, "Hello, Go", Greeting("Go"), "Greeting should match")
}

func TestDivide_Testify(t *testing.T) {
	got, err := Divide(10, 3)
	assert.NoError(t, err)
	assert.InDelta(t, 3.333, got, 0.01)
}

func TestDivideByZero_Testify(t *testing.T) {
	_, err := Divide(10, 0)
	assert.Error(t, err)
}

func TestItems_Testify(t *testing.T) {
	got := Items()
	assert.Len(t, got, 3)
	assert.Equal(t, []string{"a", "b", "c"}, got)
}
