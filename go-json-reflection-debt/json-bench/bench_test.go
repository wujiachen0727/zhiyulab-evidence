package jsonbench

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/bytedance/sonic"
	"github.com/mailru/easyjson"
)

// Order 模拟真实业务 struct：30 字段，混合类型
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

// easyjson 需要生成代码才能发挥优势，这里只对比 fallback 路径
// 为公平起见，easyjson 使用其 MarshalJSON 接口（如果有生成代码）
// 本 benchmark 的重点是 allocs/op，而非 ns/op

var testOrder = Order{
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

// v1 标准库
func BenchmarkMarshal_StdV1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(&testOrder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_StdV1(b *testing.B) {
	data, _ := json.Marshal(&testOrder)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var o Order
		if err := json.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}

// jsoniter（号称 100% 兼容 encoding/json 的 drop-in 替换）
var jsoniterAPI = jsoniter.ConfigCompatibleWithStandardLibrary

func BenchmarkMarshal_Jsoniter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := jsoniterAPI.Marshal(&testOrder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Jsoniter(b *testing.B) {
	data, _ := jsoniterAPI.Marshal(&testOrder)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var o Order
		if err := jsoniterAPI.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}

// sonic（bytedance，JIT + SIMD）
func BenchmarkMarshal_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(&testOrder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Sonic(b *testing.B) {
	data, _ := sonic.Marshal(&testOrder)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var o Order
		if err := sonic.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}

// easyjson（fallback 路径，无生成代码时退化为 encoding/json）
// 为了公平展示 easyjson 的真正优势，我们用 easyjson 的 RawMessage 方式
func BenchmarkMarshal_EasyjsonFallback(b *testing.B) {
	// easyjson 无生成代码时会 fallback，这里用 easyjson.Marshal 需要实现接口
	// 直接用 encoding/json 的结果标注为 "easyjson(fallback)" 即可说明问题
	_ = easyjson.MarshalerUnmarshaler(nil) // 编译检查
	b.Skip("easyjson 需要代码生成才能真正发挥——fallback 等同 std，跳过此项")
}
