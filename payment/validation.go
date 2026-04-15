package payment

import "fmt"

// ValidateOrder validates an order for payment operations.
func ValidateOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}
	if order.OrderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	return nil
}

// ValidateOrderID validates an order ID.
func ValidateOrderID(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	return nil
}

// ValidateRefundRequest validates a refund request.
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
