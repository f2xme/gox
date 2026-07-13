package alipay

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/f2xme/gox/payment"
)

var alipayLocation = time.FixedZone("CST", 8*60*60)

func centsToYuan(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func yuanToCents(value string) (int64, error) {
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	yuan, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || yuan < 0 {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 {
		return 0, fmt.Errorf("invalid amount precision %q", value)
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	fen, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	return yuan*100 + fen, nil
}

func mapPaymentStatus(status string) (payment.PaymentStatus, error) {
	switch status {
	case "WAIT_BUYER_PAY":
		return payment.PaymentStatusPending, nil
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return payment.PaymentStatusSuccess, nil
	case "TRADE_CLOSED":
		return payment.PaymentStatusClosed, nil
	default:
		return "", fmt.Errorf("%w: alipay %q", payment.ErrUnknownStatus, status)
	}
}

func parseAlipayTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", value, alipayLocation)
	if err != nil {
		return nil, fmt.Errorf("parse alipay time: %w", err)
	}
	return &t, nil
}
