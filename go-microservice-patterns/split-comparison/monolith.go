// 单体版：下单扣库存，一个函数搞定
// 这是一个 10 人团队的典型 Go 单体写法
package monolith

import (
	"errors"
	"sync"
)

// 内存模拟数据库
var (
	inventory = map[string]int{"sku-001": 100, "sku-002": 50}
	orders    = []Order{}
	mu        sync.Mutex
)

type Order struct {
	ID    string
	SKU   string
	Qty   int
	Total float64
}

// PlaceOrder 下单扣库存——一个事务，一个函数
func PlaceOrder(sku string, qty int, price float64) (order Order, err error) {
	mu.Lock()
	defer mu.Unlock()

	// 1. 检查库存
	stock, ok := inventory[sku]
	if !ok {
		err = errors.New("商品不存在")
		return
	}
	if stock < qty {
		err = errors.New("库存不足")
		return
	}

	// 2. 扣库存
	inventory[sku] -= qty

	// 3. 创建订单
	order = Order{
		ID:    "ord-001",
		SKU:   sku,
		Qty:   qty,
		Total: float64(qty) * price,
	}
	orders = append(orders, order)

	return
}
