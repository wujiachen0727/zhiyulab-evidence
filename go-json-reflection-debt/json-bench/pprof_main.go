package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// 同 benchmark 的 30 字段 struct
type Order struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	MerchantID    int64   `json:"merchant_id"`
	OrderNo       string  `json:"order_no"`
	Status        int     `json:"status"`
	PaymentMethod string  `json:"payment_method"`
	Currency      string  `json:"currency"`
	TotalAmount   float64 `json:"total_amount"`
	DiscountAmt   float64 `json:"discount_amt"`
	TaxAmount     float64 `json:"tax_amount"`
	ShippingFee   float64 `json:"shipping_fee"`
	FinalAmount   float64 `json:"final_amount"`
	ItemCount     int     `json:"item_count"`
	Weight        float64 `json:"weight"`
	Note          string  `json:"note"`
	ClientIP      string  `json:"client_ip"`
	UserAgent     string  `json:"user_agent"`
	Channel       string  `json:"channel"`
	Platform      string  `json:"platform"`
	DeviceID      string  `json:"device_id"`
	SessionID     string  `json:"session_id"`
	TraceID       string  `json:"trace_id"`
	RefOrderNo    string  `json:"ref_order_no"`
	CouponCode    string  `json:"coupon_code"`
	ShipAddr      string  `json:"ship_addr"`
	ShipCity      string  `json:"ship_city"`
	ShipState     string  `json:"ship_state"`
	ShipZip       string  `json:"ship_zip"`
	ShipCountry   string  `json:"ship_country"`
	CreatedAt     string  `json:"created_at"`
}

func main() {
	order := &Order{
		ID: 100001, UserID: 88123, MerchantID: 5012,
		OrderNo: "ORD-2026-05-31-00001", Status: 2,
		PaymentMethod: "wechat_pay", Currency: "CNY",
		TotalAmount: 599.00, DiscountAmt: 50.00,
		TaxAmount: 0, ShippingFee: 12.00, FinalAmount: 561.00,
		ItemCount: 3, Weight: 2.5,
		Note:      "请在工作日送货",
		ClientIP:  "192.168.1.100", UserAgent: "Mozilla/5.0 (iPhone; iOS 17)",
		Channel: "mini_program", Platform: "ios", DeviceID: "A1B2C3D4",
		SessionID: "sess_abc123", TraceID: "trace_xyz789",
		RefOrderNo: "", CouponCode: "SUMMER50",
		ShipAddr: "朝阳区望京街道1号楼", ShipCity: "北京",
		ShipState: "北京", ShipZip: "100102", ShipCountry: "CN",
		CreatedAt: "2026-05-31T10:30:00+08:00",
	}

	// CPU profile
	cpuFile, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(cpuFile)

	// 模拟高 QPS：10 秒内持续序列化 + 反序列化
	data, _ := json.Marshal(order)
	start := time.Now()
	ops := 0
	for time.Since(start) < 10*time.Second {
		json.Marshal(order)
		var o Order
		json.Unmarshal(data, &o)
		ops++
	}
	pprof.StopCPUProfile()
	cpuFile.Close()

	// Heap profile
	runtime.GC()
	heapFile, _ := os.Create("heap.prof")
	pprof.WriteHeapProfile(heapFile)
	heapFile.Close()

	fmt.Printf("完成 %d 轮 Marshal+Unmarshal（10s）\n", ops)
	fmt.Printf("CPU profile → cpu.prof\n")
	fmt.Printf("Heap profile → heap.prof\n")
}
