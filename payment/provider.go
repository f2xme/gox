package payment

import (
	"fmt"
	"strings"
)

// ParseProvider 规范化支付渠道名，供业务装配 switch 使用。
//
// 支持：mock / wechat / alipay（大小写不敏感，忽略首尾空白）。
// 空字符串视为 mock（本地默认）。未知渠道返回 ErrInvalidConfig。
//
// 生产代码可只依赖本包解析渠道，再分别装配 mock / wechat / alipay adapter，
// 无需为解析字符串而 import 测试用 mock 包。
func ParseProvider(name string) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "mock":
		return ProviderMock, nil
	case "wechat", "wx", "weixin":
		return ProviderWechat, nil
	case "alipay", "ali":
		return ProviderAlipay, nil
	default:
		return "", fmt.Errorf("%w: unknown payment provider %q (want mock, wechat or alipay)", ErrInvalidConfig, name)
	}
}
