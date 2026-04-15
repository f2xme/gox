package alipay

import (
	"testing"

	"github.com/f2xme/gox/payment"
)

// TestAlipayImplementsPaymentInterface verifies that Alipay implements payment.Payment interface.
func TestAlipayImplementsPaymentInterface(t *testing.T) {
	var _ payment.Payment = (*Alipay)(nil)
}

// TestAlipay_Pay tests the Pay method.
func TestAlipay_Pay(t *testing.T) {
	tests := []struct {
		name    string
		order   *payment.Order
		wantErr bool
	}{
		{
			name: "valid order",
			order: &payment.Order{
				OrderID:     "TEST001",
				Amount:      10000,
				Subject:     "测试商品",
				Description: "测试订单",
				NotifyURL:   "https://example.com/notify",
				ReturnURL:   "https://example.com/return",
			},
			wantErr: false,
		},
	}

	ap := NewAlipay("test_appid", "test_private_key", "test_public_key", true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ap.Pay(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == nil {
					t.Error("Pay() result is nil")
					return
				}
				if result.OrderID != tt.order.OrderID {
					t.Errorf("Pay() OrderID = %v, want %v", result.OrderID, tt.order.OrderID)
				}
				if result.TransactionID == "" {
					t.Error("Pay() TransactionID is empty")
				}
				if result.PayURL == "" {
					t.Error("Pay() PayURL is empty")
				}
			}
		})
	}
}

// TestAlipay_Query tests the Query method.
func TestAlipay_Query(t *testing.T) {
	tests := []struct {
		name    string
		orderID string
		wantErr bool
	}{
		{
			name:    "valid order ID",
			orderID: "TEST001",
			wantErr: false,
		},
	}

	ap := NewAlipay("test_appid", "test_private_key", "test_public_key", true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ap.Query(tt.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == nil {
					t.Error("Query() result is nil")
					return
				}
				if result.OrderID != tt.orderID {
					t.Errorf("Query() OrderID = %v, want %v", result.OrderID, tt.orderID)
				}
				if result.Status == "" {
					t.Error("Query() Status is empty")
				}
			}
		})
	}
}

// TestAlipay_Refund tests the Refund method.
func TestAlipay_Refund(t *testing.T) {
	tests := []struct {
		name    string
		req     *payment.RefundRequest
		wantErr bool
	}{
		{
			name: "valid refund request",
			req: &payment.RefundRequest{
				OrderID:  "TEST001",
				RefundID: "REFUND001",
				Amount:   5000,
				Reason:   "测试退款",
			},
			wantErr: false,
		},
	}

	ap := NewAlipay("test_appid", "test_private_key", "test_public_key", true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ap.Refund(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Refund() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == nil {
					t.Error("Refund() result is nil")
					return
				}
				if result.RefundID != tt.req.RefundID {
					t.Errorf("Refund() RefundID = %v, want %v", result.RefundID, tt.req.RefundID)
				}
				if result.Status == "" {
					t.Error("Refund() Status is empty")
				}
			}
		})
	}
}

// TestAlipay_Close tests the Close method.
func TestAlipay_Close(t *testing.T) {
	tests := []struct {
		name    string
		orderID string
		wantErr bool
	}{
		{
			name:    "valid order ID",
			orderID: "TEST001",
			wantErr: false,
		},
	}

	ap := NewAlipay("test_appid", "test_private_key", "test_public_key", true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ap.Close(tt.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
