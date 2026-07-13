package wechat

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/f2xme/gox/payment"
	wx "github.com/go-pay/gopay/wechat/v3"
)

func validateCall(ctx context.Context, order *payment.Order) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	return payment.ValidateOrder(order)
}

func mapPaymentStatus(value string) (payment.PaymentStatus, error) {
	switch value {
	case "SUCCESS":
		return payment.PaymentStatusSuccess, nil
	case "REFUND":
		return payment.PaymentStatusRefunded, nil
	case "NOTPAY", "USERPAYING", "ACCEPT":
		return payment.PaymentStatusPending, nil
	case "CLOSED", "REVOKED":
		return payment.PaymentStatusClosed, nil
	case "PAYERROR":
		return payment.PaymentStatusFailed, nil
	default:
		return "", fmt.Errorf("%w: wechat %q", payment.ErrUnknownStatus, value)
	}
}

func mapRefundStatus(value string) (payment.RefundStatus, error) {
	switch value {
	case "SUCCESS":
		return payment.RefundStatusSuccess, nil
	case "PROCESSING":
		return payment.RefundStatusPending, nil
	case "CLOSED":
		return payment.RefundStatusClosed, nil
	case "ABNORMAL":
		return payment.RefundStatusFailed, nil
	default:
		return "", fmt.Errorf("%w: wechat refund %q", payment.ErrUnknownStatus, value)
	}
}

func parseWechatTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse wechat time: %w", err)
	}
	return &t, nil
}

func checkResponse(code int, e wx.ErrResponse, hasBody bool) error {
	if code != wx.Success {
		return &gatewayResponseError{Code: e.Code, Message: e.Message, HTTPCode: code}
	}
	if !hasBody {
		return fmt.Errorf("empty gateway response")
	}
	return nil
}

type gatewayResponseError struct {
	Code, Message string
	HTTPCode      int
}

func (e *gatewayResponseError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return "HTTP " + strconv.Itoa(e.HTTPCode)
}

func providerError(operation string, err error) error {
	if err == nil {
		err = payment.ErrGateway
	}
	e := &payment.ProviderError{Provider: payment.ProviderWechat, Operation: operation, Err: fmt.Errorf("%w: %w", payment.ErrGateway, err)}
	if responseErr, ok := err.(*gatewayResponseError); ok {
		e.Code, e.Message = responseErr.Code, responseErr.Message
	}
	return e
}

func respCode(value any) int {
	switch r := value.(type) {
	case *wx.NativeRsp:
		if r != nil {
			return r.Code
		}
	case *wx.PrepayRsp:
		if r != nil {
			return r.Code
		}
	case *wx.QueryOrderRsp:
		if r != nil {
			return r.Code
		}
	case *wx.RefundRsp:
		if r != nil {
			return r.Code
		}
	case *wx.EmptyRsp:
		if r != nil {
			return r.Code
		}
	}
	return -1
}
func nativeError(r *wx.NativeRsp) wx.ErrResponse {
	if r != nil {
		return r.ErrResponse
	}
	return wx.ErrResponse{}
}
func prepayError(r *wx.PrepayRsp) wx.ErrResponse {
	if r != nil {
		return r.ErrResponse
	}
	return wx.ErrResponse{}
}
func queryError(r *wx.QueryOrderRsp) wx.ErrResponse {
	if r != nil {
		return r.ErrResponse
	}
	return wx.ErrResponse{}
}
func refundError(r *wx.RefundRsp) wx.ErrResponse {
	if r != nil {
		return r.ErrResponse
	}
	return wx.ErrResponse{}
}
func emptyError(r *wx.EmptyRsp) wx.ErrResponse {
	if r != nil {
		return r.ErrResponse
	}
	return wx.ErrResponse{}
}
