package ip2region

import (
	"context"
	"strings"
	"unicode"

	"github.com/lionsoul2014/ip2region/binding/golang/service"

	"github.com/f2xme/gox/geo"
)

// searcher 抽象底层查询，便于单测注入。
//
// 签名与 service.Ip2Region.Search 保持一致（参数为 any）。
type searcher interface {
	Search(ip any) (string, error)
	Close()
}

// Locator 是基于 ip2region 的离线 IP 地区查询实现。
type Locator struct {
	service searcher
	options Options
}

var _ geo.Locator = (*Locator)(nil)

// Lookup 查询 IP 对应的地区信息。
//
// 注意：底层 xdb Search 为同步阻塞调用，仅在入口检查 ctx 是否已取消/超时；
// 查询进行中无法被 context 中途打断。需要严格超时请在调用侧控制，或使用
// BufferCache 以降低磁盘 IO 耗时。
func (l *Locator) Lookup(ctx context.Context, ip string) (*geo.Location, error) {
	if ctx == nil {
		return nil, geo.NewError(geo.ErrCodeInvalidArgument, "context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "context error", err, ip)
	}

	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return nil, err
	}

	region, err := l.service.Search(normalized)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "search failed", err, normalized)
	}
	region = strings.TrimSpace(region)
	if region == "" {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}

	loc := parseRegion(region)
	loc.IP = normalized
	if loc.Empty() || isPrivateRegion(loc) {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	return loc, nil
}

// Close 关闭底层查询服务并释放资源。
func (l *Locator) Close() error {
	if l == nil || l.service == nil {
		return nil
	}
	l.service.Close()
	return nil
}

// parseRegion 解析 ip2region 返回的 region 字符串。
//
// 兼容两种常见 5 段格式：
//   - 新格式：国家|省|市|ISP|国家代码
//   - 旧格式：国家|区域|省|市|ISP
func parseRegion(region string) *geo.Location {
	parts := strings.Split(region, "|")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "0" {
			p = ""
		}
		parts[i] = p
	}

	loc := &geo.Location{}
	switch len(parts) {
	case 5:
		if isCountryCode(parts[4]) {
			// 新格式：国家|省|市|ISP|国家代码
			loc.Country = parts[0]
			loc.Province = parts[1]
			loc.City = parts[2]
			loc.ISP = parts[3]
			loc.CountryCode = strings.ToUpper(parts[4])
		} else {
			// 旧格式：国家|区域|省|市|ISP
			loc.Country = parts[0]
			// parts[1] 为区域，多数情况为 0，忽略
			loc.Province = parts[2]
			loc.City = parts[3]
			loc.ISP = parts[4]
		}
	case 4:
		loc.Country = parts[0]
		loc.Province = parts[1]
		loc.City = parts[2]
		loc.ISP = parts[3]
	case 3:
		loc.Country = parts[0]
		loc.Province = parts[1]
		loc.City = parts[2]
	case 2:
		loc.Country = parts[0]
		loc.Province = parts[1]
	case 1:
		loc.Country = parts[0]
	default:
		if len(parts) > 5 {
			// 超长时按新格式前 5 段尽量解析
			loc.Country = parts[0]
			loc.Province = parts[1]
			loc.City = parts[2]
			loc.ISP = parts[3]
			if isCountryCode(parts[4]) {
				loc.CountryCode = strings.ToUpper(parts[4])
			}
		}
	}

	return loc
}

// isPrivateRegion 判断是否为内网/保留地址占位结果。
func isPrivateRegion(loc *geo.Location) bool {
	if loc == nil {
		return false
	}
	for _, v := range []string{loc.Country, loc.Province, loc.City, loc.ISP} {
		if strings.Contains(v, "内网IP") || strings.EqualFold(v, "局域网") {
			return true
		}
	}
	return false
}

func isCountryCode(s string) bool {
	if len(s) != 2 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// 确保 *service.Ip2Region 满足 searcher（编译期检查，实际类型在 new.go 注入）。
var _ searcher = (*service.Ip2Region)(nil)
