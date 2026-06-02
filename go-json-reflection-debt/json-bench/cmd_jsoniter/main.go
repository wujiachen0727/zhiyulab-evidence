package main

import (
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
)

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

var api = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	o := Order{ID: 1, OrderNo: "test"}
	data, _ := api.Marshal(&o)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}
