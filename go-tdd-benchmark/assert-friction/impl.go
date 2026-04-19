package assertfriction

import "fmt"

func Add(a, b int) int { return a + b }
func Greeting(name string) string { return "Hello, " + name }
func Divide(a, b float64) (float64, error) {
	if b == 0 { return 0, fmt.Errorf("division by zero") }
	return a / b, nil
}
func Items() []string { return []string{"a", "b", "c"} }
