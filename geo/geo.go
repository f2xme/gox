package geo

import (
	"context"
	"net"
	"strings"
)

// Locator 定义 IP 地区查询能力。
//
// 实现应当是并发安全的。若实现持有需要释放的资源，可额外实现 Close 方法，
// 但不强制要求所有实现都可关闭。
type Locator interface {
	// Lookup 根据 IP 地址查询地理位置信息。
	//
	// ip 可以是 IPv4 或 IPv6 字符串。返回的 Location 至少应填充 IP 字段。
	// 查询失败时返回错误；IP 格式无效时应返回 IsInvalidIP 可识别的错误。
	Lookup(ctx context.Context, ip string) (*Location, error)
}

// Location 表示 IP 对应的地理位置信息。
//
// 不同数据源字段完整度不同，未提供的字段保持零值。
type Location struct {
	// IP 查询的原始 IP 地址。
	IP string
	// Country 国家名称。
	Country string
	// CountryCode 国家/地区代码，例如 CN、US。
	CountryCode string
	// Province 一级行政区（省/州）。
	Province string
	// City 城市。
	City string
	// District 区/县。
	District string
	// ISP 网络运营商。
	ISP string
	// Latitude 纬度，未提供时为 0。
	Latitude float64
	// Longitude 经度，未提供时为 0。
	Longitude float64
	// Extra 实现方可附带的扩展字段。
	Extra map[string]string
}

// String 返回便于日志输出的地区摘要。
func (l *Location) String() string {
	if l == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	for _, part := range []string{l.Country, l.Province, l.City, l.ISP} {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return l.IP
	}
	return strings.Join(parts, " ")
}

// Empty 判断是否几乎没有任何地区信息。
//
// 仅有 IP 字段时视为空结果。
func (l *Location) Empty() bool {
	if l == nil {
		return true
	}
	return strings.TrimSpace(l.Country) == "" &&
		strings.TrimSpace(l.CountryCode) == "" &&
		strings.TrimSpace(l.Province) == "" &&
		strings.TrimSpace(l.City) == "" &&
		strings.TrimSpace(l.District) == "" &&
		strings.TrimSpace(l.ISP) == "" &&
		l.Latitude == 0 &&
		l.Longitude == 0 &&
		len(l.Extra) == 0
}

// Clone 返回 Location 的深拷贝。
func (l *Location) Clone() *Location {
	if l == nil {
		return nil
	}
	cloned := *l
	if len(l.Extra) > 0 {
		cloned.Extra = make(map[string]string, len(l.Extra))
		for k, v := range l.Extra {
			cloned.Extra[k] = v
		}
	}
	return &cloned
}

// NormalizeIP 校验并规范化 IP 地址字符串。
//
// 成功时返回规范化后的 IP（IPv4 点分十进制或 IPv6 压缩格式），
// 失败时返回 ErrCodeInvalidIP 错误。
func NormalizeIP(ip string) (string, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "", NewError(ErrCodeInvalidIP, "ip is required")
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", NewError(ErrCodeInvalidIP, "invalid ip address: "+ip)
	}
	// 优先输出 IPv4 点分格式
	if v4 := parsed.To4(); v4 != nil {
		return v4.String(), nil
	}
	return parsed.String(), nil
}
