package payment

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ValidateContext 校验调用上下文。
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("%w: context cannot be nil", ErrInvalidRequest)
	}
	return nil
}

// ValidateOrder 校验支付订单。
func ValidateOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("%w: order cannot be nil", ErrInvalidRequest)
	}
	if order.OrderID == "" {
		return fmt.Errorf("%w: order ID cannot be empty", ErrInvalidRequest)
	}
	if order.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidRequest)
	}
	if order.Subject == "" {
		return fmt.Errorf("%w: subject cannot be empty", ErrInvalidRequest)
	}
	if err := validateHTTPURL("notify URL", order.NotifyURL, true); err != nil {
		return err
	}
	if err := validateHTTPURL("return URL", order.ReturnURL, false); err != nil {
		return err
	}
	if order.ExpireAt != nil && !order.ExpireAt.After(time.Now()) {
		return fmt.Errorf("%w: expire time must be in the future", ErrInvalidRequest)
	}
	return nil
}

// ValidateOrderID 校验商户订单号。
func ValidateOrderID(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("%w: order ID cannot be empty", ErrInvalidRequest)
	}
	return nil
}

// ValidateRefundRequest 校验退款请求。
func ValidateRefundRequest(req *RefundRequest) error {
	if req == nil {
		return fmt.Errorf("%w: refund request cannot be nil", ErrInvalidRequest)
	}
	if req.OrderID == "" {
		return fmt.Errorf("%w: order ID cannot be empty", ErrInvalidRequest)
	}
	if req.RefundID == "" {
		return fmt.Errorf("%w: refund ID cannot be empty", ErrInvalidRequest)
	}
	if req.Amount <= 0 {
		return fmt.Errorf("%w: refund amount must be positive", ErrInvalidRequest)
	}
	if req.OriginalAmount <= 0 {
		return fmt.Errorf("%w: original amount must be positive", ErrInvalidRequest)
	}
	if req.Amount > req.OriginalAmount {
		return fmt.Errorf("%w: refund amount exceeds original amount", ErrInvalidRequest)
	}
	return validateHTTPURL("notify URL", req.NotifyURL, false)
}

func validateHTTPURL(name, raw string, required bool) error {
	if raw == "" {
		if required {
			return fmt.Errorf("%w: %s cannot be empty", ErrInvalidRequest, name)
		}
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("%w: %s must be an absolute HTTP URL", ErrInvalidRequest, name)
	}
	return nil
}
