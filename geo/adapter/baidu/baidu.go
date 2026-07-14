package baidu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/f2xme/gox/geo"
)

// Locator 是基于百度 IP 查询接口的实现。
type Locator struct {
	endpoint string
	client   *http.Client
}

var _ geo.Locator = (*Locator)(nil)

type baiduResponse struct {
	// Status 兼容字符串 "0" 与数字 0。
	Status flexibleString `json:"status"`
	Data   []struct {
		Location    string `json:"location"`
		OrigIP      string `json:"origip"`
		OrigIPQuery string `json:"origipquery"`
	} `json:"data"`
}

// flexibleString 兼容 JSON 中的 string / number 状态字段。
type flexibleString string

// UnmarshalJSON 实现 json.Unmarshaler。
func (s *flexibleString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*s = ""
		return nil
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = flexibleString(str)
		return nil
	}
	// 数字或其他标量：去掉引号后按原始文本保存
	*s = flexibleString(string(data))
	return nil
}

// Lookup 通过百度接口查询 IP 地区信息。
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

	reqURL, err := url.Parse(l.endpoint)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInvalidArgument, "invalid endpoint", err, normalized)
	}
	query := reqURL.Query()
	// 保留调用方自定义 query，并覆盖接口所需参数
	query.Set("resource_id", "6006")
	query.Set("format", "json")
	query.Set("tn", "baidu")
	query.Set("query", normalized)
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "create request failed", err, normalized)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeUnavailable, "http request failed", err, normalized)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "read response failed", err, normalized)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, geo.NewError(
			geo.ErrCodeUnavailable,
			fmt.Sprintf("unexpected status code %d", resp.StatusCode),
			normalized,
		)
	}

	text, err := decodeResponseBody(body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "decode response failed", err, normalized)
	}

	var raw baiduResponse
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "parse response failed", err, normalized)
	}

	// status 为 "0" 或 0 表示成功；部分场景也会直接返回 data
	status := strings.TrimSpace(string(raw.Status))
	if status != "" && status != "0" {
		return nil, geo.NewError(geo.ErrCodeUnavailable, "baidu lookup failed: status="+status, normalized)
	}
	if len(raw.Data) == 0 {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}

	locationText := strings.TrimSpace(raw.Data[0].Location)
	if locationText == "" {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	if isPrivateLocationText(locationText) {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}

	loc := parseBaiduLocation(locationText)
	loc.IP = normalized
	if loc.Extra == nil {
		loc.Extra = make(map[string]string)
	}
	loc.Extra["location"] = locationText
	if raw.Data[0].OrigIP != "" {
		loc.Extra["origip"] = raw.Data[0].OrigIP
	}
	return loc, nil
}

// decodeResponseBody 将响应体解码为 UTF-8 字符串。
//
// 优先根据 Content-Type 的 charset 选择解码方式；未知时若已是合法 UTF-8 则直接使用，否则按 GBK 解码。
func decodeResponseBody(body []byte, contentType string) (string, error) {
	charset := parseCharset(contentType)
	switch charset {
	case "gbk", "gb2312", "gb18030":
		return decodeGBK(body)
	case "utf-8", "utf8":
		return string(body), nil
	default:
		if utf8.Valid(body) {
			return string(body), nil
		}
		return decodeGBK(body)
	}
}

func parseCharset(contentType string) string {
	contentType = strings.ToLower(contentType)
	const marker = "charset="
	idx := strings.Index(contentType, marker)
	if idx < 0 {
		return ""
	}
	rest := contentType[idx+len(marker):]
	rest = strings.TrimSpace(rest)
	// 去掉可能的引号与后续参数
	rest = strings.Trim(rest, `"'`)
	if i := strings.IndexAny(rest, " ;,"); i >= 0 {
		rest = rest[:i]
	}
	return strings.TrimSpace(rest)
}

func decodeGBK(body []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// isPrivateLocationText 识别内网/保留地址等占位文案。
func isPrivateLocationText(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	// 去掉空白后匹配常见占位
	compact := strings.ReplaceAll(t, " ", "")
	markers := []string{
		"内网IP", "内网ip", "局域网", "本机地址", "本机", "保留地址",
		"本地", "回环", "私有地址", "专用网络",
	}
	for _, m := range markers {
		if strings.Contains(compact, m) || strings.EqualFold(compact, m) {
			return true
		}
	}
	lower := strings.ToLower(t)
	for _, m := range []string{"private", "reserved", "localhost", "loopback", "bogon"} {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// parseBaiduLocation 解析类似 "广东省深圳市 电信" / "北京市 联通" / "美国" 的文本。
//
// 国内：命中省/自治区/直辖市/特别行政区时填 Country=中国。
// 海外：首段作 Country；其余段保守放入 Extra 原文，仅在像行政区时拆 Province。
// 单独「xx市」不再默认判定为中国（避免「大阪市」误标 CN）。
func parseBaiduLocation(text string) *geo.Location {
	loc := &geo.Location{}

	fields := strings.Fields(text)
	if len(fields) == 0 {
		return loc
	}

	region := fields[0]

	// 国内行政区 → Country=中国；海外文本 → Country=首段
	if isChineseAdminRegion(region) {
		loc.Country = "中国"
		loc.CountryCode = "CN"
		province, city, district := splitChineseRegion(region)
		loc.Province = province
		loc.City = city
		loc.District = district
		if len(fields) > 1 {
			loc.ISP = strings.Join(fields[1:], " ")
		}
		return loc
	}

	// 海外：首段为国家；无法可靠识别 ISP 时把后续段落放 Province（若像行政区），
	// 否则整段落进 ISP，完整原文由调用方 Extra["location"] 兜底。
	loc.Country = region
	if len(fields) == 1 {
		return loc
	}
	rest := fields[1:]
	if looksLikeISP(rest[0]) {
		loc.ISP = strings.Join(rest, " ")
		return loc
	}
	// 第二段像行政区名称时作为 Province，再后面若像 ISP 则填 ISP
	loc.Province = rest[0]
	if len(rest) > 1 {
		loc.ISP = strings.Join(rest[1:], " ")
	}
	return loc
}

func looksLikeISP(s string) bool {
	lower := strings.ToLower(s)
	// 国内运营商关键词
	for _, kw := range []string{"电信", "联通", "移动", "铁通", "广电", "教育网", "科技网", "鹏博士", "长城", "有线通"} {
		if strings.Contains(s, kw) {
			return true
		}
	}
	// 常见海外运营商/组织片段（保守，宁可少匹配）
	for _, kw := range []string{"telecom", "mobile", "unicom", "broadband", "communications", "ltd", "inc", "corp", "at&t", "softbank", "verizon", "comcast"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func isChineseAdminRegion(region string) bool {
	region = strings.TrimSpace(region)
	if region == "" {
		return false
	}
	// 省 / 自治区 / 特别行政区：明确国内结构
	if strings.Contains(region, "省") ||
		strings.Contains(region, "自治区") ||
		strings.Contains(region, "特别行政区") {
		return true
	}
	// 直辖市（含区县后缀）
	for _, name := range []string{"北京市", "上海市", "天津市", "重庆市"} {
		if strings.HasPrefix(region, name) {
			return true
		}
	}
	// 不把单独的「xx市」判为中国，避免「大阪市」「首尔市」误标 CN
	return false
}

// splitChineseRegion 从 "广东省深圳市" / "北京市" / "重庆市渝中区" 拆分省、市、区。
func splitChineseRegion(region string) (province, city, district string) {
	region = strings.TrimSpace(region)
	if region == "" {
		return "", "", ""
	}

	// 自治区
	if idx := strings.Index(region, "自治区"); idx >= 0 {
		end := idx + len("自治区")
		province = region[:end]
		rest := region[end:]
		city, district = splitCityDistrict(rest)
		return province, city, district
	}
	// 特别行政区
	if idx := strings.Index(region, "特别行政区"); idx >= 0 {
		end := idx + len("特别行政区")
		province = region[:end]
		rest := region[end:]
		if rest == "" {
			return province, province, ""
		}
		city, district = splitCityDistrict(rest)
		return province, city, district
	}
	// 省
	if idx := strings.Index(region, "省"); idx >= 0 {
		end := idx + len("省")
		province = region[:end]
		rest := region[end:]
		city, district = splitCityDistrict(rest)
		return province, city, district
	}
	// 直辖市：北京市 / 上海市 / 天津市 / 重庆市（可带区县后缀）
	for _, name := range []string{"北京市", "上海市", "天津市", "重庆市"} {
		if strings.HasPrefix(region, name) {
			rest := region[len(name):]
			if rest == "" {
				return name, name, ""
			}
			// 后缀作为区县
			return name, name, rest
		}
	}
	// 仅 "xx市"
	if strings.HasSuffix(region, "市") {
		return region, region, ""
	}
	return region, "", ""
}

// splitCityDistrict 从 "深圳市" / "深圳市南山区" 拆分市与区。
func splitCityDistrict(rest string) (city, district string) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", ""
	}
	if idx := strings.Index(rest, "市"); idx >= 0 {
		end := idx + len("市")
		city = rest[:end]
		district = rest[end:]
		return city, district
	}
	// 无“市”时整段作为 city
	return rest, ""
}
