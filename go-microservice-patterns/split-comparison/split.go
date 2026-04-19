// 拆分版：同样的"下单扣库存"，拆成订单服务 + 库存服务
// 看看代码膨胀了多少
package split

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ========== 库存服务（Inventory Service） ==========

type DeductRequest struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type DeductResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Remain  int    `json:"remain"`
}

// ========== 订单服务（Order Service） ==========

type Order struct {
	ID    string
	SKU   string
	Qty   int
	Total float64
}

// InventoryClient 库存服务的 HTTP 客户端
type InventoryClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second, // 超时控制——单体版不需要
		},
	}
}

// DeductStock 扣库存——原来是一行代码，现在是一次网络调用
func (c *InventoryClient) DeductStock(ctx context.Context, sku string, qty int) (remain int, err error) {
	// 1. 构造请求
	reqBody := DeductRequest{SKU: sku, Qty: qty}
	data, err := json.Marshal(reqBody)
	if err != nil {
		err = fmt.Errorf("序列化失败: %w", err)
		return
	}

	// 2. 发送 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/deduct", nil)
	if err != nil {
		err = fmt.Errorf("构造请求失败: %w", err)
		return
	}
	_ = data // 简化示例，实际需要设置 body

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// 网络错误——单体版根本不存在这类问题
		err = fmt.Errorf("库存服务不可达: %w", err)
		return
	}
	defer resp.Body.Close()

	// 3. 解析响应
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("库存服务返回异常: status=%d", resp.StatusCode)
		return
	}

	var deductResp DeductResponse
	if err = json.NewDecoder(resp.Body).Decode(&deductResp); err != nil {
		err = fmt.Errorf("解析响应失败: %w", err)
		return
	}

	if !deductResp.Success {
		err = errors.New(deductResp.Error)
		return
	}

	remain = deductResp.Remain
	return
}

// PlaceOrder 下单——原来 20 行，现在光错误处理就 20 行
func PlaceOrder(ctx context.Context, invClient *InventoryClient, sku string, qty int, price float64) (order Order, err error) {
	// 1. 调用库存服务扣库存（跨网络）
	_, err = invClient.DeductStock(ctx, sku, qty)
	if err != nil {
		// 问题来了：扣库存失败，要不要重试？
		// 重试的话，万一第一次其实成功了呢？（幂等性问题）
		// 不重试的话，可能只是网络抖动
		err = fmt.Errorf("扣库存失败: %w", err)
		return
	}

	// 2. 创建订单
	order = Order{
		ID:    "ord-001",
		SKU:   sku,
		Qty:   qty,
		Total: float64(qty) * price,
	}

	// 问题又来了：如果这里创建订单失败，库存已经扣了怎么办？
	// 单体版：一个事务回滚就行
	// 拆分版：需要补偿事务（Saga）或两阶段提交（2PC）
	// 这又是几百行代码...

	return
}
