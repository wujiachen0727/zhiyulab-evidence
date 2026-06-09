// E1: Java in Go 的反面示例 — 用 Go 模拟继承的典型代码
// 这是很多从 Java 转 Go 的开发者在项目初期容易写出的代码

package main

import "fmt"

// ---- 反面示例：用 Go 模拟 Java 的类继承 ----

// Animal 是一个"基类"
type Animal struct {
    Name string
}

func (a *Animal) Speak() string {
    return "..."
}

func (a *Animal) GetName() string {
    return a.Name
}

// Dog "继承" Animal
type Dog struct {
    Animal // 嵌入
    Breed string
}

func (d *Dog) Speak() string {
    return "Woof!"
}

// Cat "继承" Animal
type Cat struct {
    Animal
    Color string
}

func (c *Cat) Speak() string {
    return "Meow!"
}

// ---- 使用方需要知道具体类型 ----
func PrintAnimalSpeak(a *Animal) {
    fmt.Printf("%s says: %s\n", a.GetName(), a.Speak())
    // 问题：这里只能调用 Animal 的 Speak()，无法体现多态
    // 因为 Go 的嵌入是组合，不是继承
}

func main() {
    dog := &Dog{Animal: Animal{Name: "Buddy"}, Breed: "Golden Retriever"}
    cat := &Cat{Animal: Animal{Name: "Kitty"}, Color: "White"}

    fmt.Println("--- Java in Go: 用嵌入模拟继承 ---")
    fmt.Printf("%s (%s) says: %s\n", dog.GetName(), dog.Breed, dog.Speak())
    fmt.Printf("%s (%s) says: %s\n", cat.GetName(), cat.Color, cat.Speak())

    // 问题 1: 无法通过 Animal 指针调用子类的 Speak()
    fmt.Println("\n--- 问题：通过 Animal 指针调用 ---")
    PrintAnimalSpeak(&dog.Animal) // 输出 "...", 不是 "Woof!"
    PrintAnimalSpeak(&cat.Animal) // 输出 "...", 不是 "Meow!"

    // 问题 2: 如果 Animal 有初始化逻辑，嵌入会导致重复初始化
    // 问题 3: 类型判断需要类型断言，编译期无法保证安全
}
