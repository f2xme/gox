package payment

import "fmt"

// ValidateOrder 校验支付订单。
func ValidateOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}
	if order.OrderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	return nil
}

// ValidateOrderID 校验商户订单号。
func ValidateOrderID(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	return nil
}

// ValidateRefundRequest 校验退款请求。
func ValidateRefundRequest(req *RefundRequest) error {
	if req == nil {
		return fmt.Errorf("refund request cannot be nil")
	}
	if req.OrderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if req.RefundID == "" {
		return fmt.Errorf("refund ID cannot be empty")
	}
	return nil
}
