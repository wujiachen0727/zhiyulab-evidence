package main

import "fmt"

func show(tag string, s []int) {
	fmt.Printf("  %-26s len=%d cap=%d 值=%v\n", tag, len(s), cap(s), s)
}

func main() {
	fmt.Println("== 组 1：len < cap，append 原地写进共享数组 ==")
	a := make([]int, 3, 4)
	a[0], a[1], a[2] = 1, 2, 3
	show("a", a)

	head := a[:2]
	show("head = a[:2]", head)

	head = append(head, 99)
	show("head 追加 99 后", head)
	show("a 被改了", a)

	fmt.Println()
	fmt.Println("== 组 2：子切片的容量继承到原数组末尾 ==")
	b := make([]int, 3, 3)
	b[0], b[1], b[2] = 1, 2, 3
	show("b", b)

	bHead := b[:2]
	fmt.Printf("  %-26s len=2 但 cap=%d（不是 2）\n", "b[:2]", cap(bHead))
	bHead = append(bHead, 99)
	show("b[:2] 追加 99 后", bHead)
	show("b 还是被改了", b)

	c := make([]int, 3, 3)
	c[0], c[1], c[2] = 1, 2, 3
	show("c", c)

	cHead := c[:2:2]
	fmt.Printf("  %-26s len=2 且 cap=%d（三下标把容量也切了）\n", "c[:2:2]", cap(cHead))
	cHead = append(cHead, 99)
	show("c[:2:2] 追加 99 后", cHead)
	show("c 这次没变", c)

	fmt.Println()
	fmt.Println("== 组 3：容量够用时连做两次 append，两次互相覆盖 ==")
	d := make([]int, 3, 4)
	d[0], d[1], d[2] = 1, 2, 3
	show("d", d)

	e := append(d, 10)
	f := append(d, 20)
	show("e = append(d, 10)", e)
	show("f = append(d, 20)", f)
	fmt.Println("  改 e[0] = 777，再看 f：")
	e[0] = 777
	show("f", f)

	fmt.Println()
	fmt.Println("== 组 4：越界扩容之后，关系断开 ==")
	g := make([]int, 3, 4)
	g[0], g[1], g[2] = 1, 2, 3
	show("g", g)

	h := append(g, 10, 20)
	show("h = append(g, 10, 20)", h)
	fmt.Println("  改 h[0] = 888，再看 g：")
	h[0] = 888
	show("g", g)
}
