package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f2xme/gox/geo"
)

// Locator 是基于 HTTP API 的 IP 地区查询实现。
type Locator struct {
	endpoint string
	client   *http.Client
	headers  map[string]string
	parser   ResponseParser
}

var _ geo.Locator = (*Locator)(nil)

// Lookup 通过 HTTP API 查询 IP 地区信息。
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

	reqURL, err := buildRequestURL(l.endpoint, normalized)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInvalidArgument, "invalid endpoint", err, normalized)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "create request failed", err, normalized)
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range l.headers {
		req.Header.Set(k, v)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeUnavailable, "http request failed", err, normalized)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "read response failed", err, normalized)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, geo.NewError(
			geo.ErrCodeUnavailable,
			fmt.Sprintf("unexpected status code %d", resp.StatusCode),
			normalized,
		)
	}

	loc, err := l.parser(body, resp.StatusCode, normalized)
	if err != nil {
		if geo.ErrorCode(err) != "" {
			return nil, err
		}
		return nil, geo.WrapError(geo.ErrCodeInternal, "parse response failed", err, normalized)
	}
	if loc == nil || loc.Empty() {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	if loc.IP == "" {
		loc.IP = normalized
	}
	return loc, nil
}

func buildRequestURL(endpoint, ip string) (string, error) {
	escaped := url.PathEscape(ip)
	// 使用 Replace 而非 Sprintf，避免 endpoint 中已有的 %XX 百分号编码被误解析。
	if strings.Contains(endpoint, "%s") {
		return strings.Replace(endpoint, "%s", escaped, 1), nil
	}
	// 前缀拼接：确保中间有分隔
	if strings.HasSuffix(endpoint, "/") || strings.HasSuffix(endpoint, "=") || strings.HasSuffix(endpoint, "?") {
		return endpoint + escaped, nil
	}
	return endpoint + "/" + escaped, nil
}

func defaultJSONParser(body []byte, statusCode int, ip string) (*geo.Location, error) {
	_ = statusCode

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	// ip-api.com 风格：status=fail
	if status, ok := stringField(raw, "status"); ok && strings.EqualFold(status, "fail") {
		msg, _ := stringField(raw, "message")
		if msg == "" {
			msg = "lookup failed"
		}
		if isNotFoundFailMessage(msg) {
			return nil, geo.NewError(geo.ErrCodeNotFound, msg, ip)
		}
		// invalid key / rate limit 等配置与上游故障 → Unavailable
		return nil, geo.NewError(geo.ErrCodeUnavailable, msg, ip)
	}

	loc := &geo.Location{IP: ip}
	loc.Country, _ = stringField(raw, "country", "country_name", "countryName")
	loc.CountryCode, _ = stringField(raw, "countryCode", "country_code", "countrycode")
	loc.Province, _ = stringField(raw, "regionName", "province", "region_name", "region")
	loc.City, _ = stringField(raw, "city")
	loc.District, _ = stringField(raw, "district", "county")
	loc.ISP, _ = stringField(raw, "isp", "org", "as", "organization")

	if lat, ok := floatField(raw, "lat", "latitude"); ok {
		loc.Latitude = lat
	}
	if lon, ok := floatField(raw, "lon", "lng", "longitude"); ok {
		loc.Longitude = lon
	}

	// 嵌套 data 对象兼容
	if loc.Empty() {
		if nested, ok := raw["data"].(map[string]any); ok {
			return defaultJSONParser(mustJSON(nested), statusCode, ip)
		}
	}

	return loc, nil
}

func mustJSON(v map[string]any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func stringField(raw map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if v, ok := raw[key]; ok {
			switch val := v.(type) {
			case string:
				if s := strings.TrimSpace(val); s != "" && s != "0" {
					return s, true
				}
			case float64:
				if val != 0 {
					return strconv.FormatFloat(val, 'f', -1, 64), true
				}
			case json.Number:
				if s := val.String(); s != "" && s != "0" {
					return s, true
				}
			}
		}
		// 大小写不敏感回退
		lowerKey := strings.ToLower(key)
		for k, v := range raw {
			if strings.ToLower(k) != lowerKey {
				continue
			}
			if s, ok := anyToString(v); ok {
				return s, true
			}
		}
	}
	return "", false
}

func floatField(raw map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		if v, ok := raw[key]; ok {
			if f, ok := anyToFloat(v); ok {
				return f, true
			}
		}
		lowerKey := strings.ToLower(key)
		for k, v := range raw {
			if strings.ToLower(k) != lowerKey {
				continue
			}
			if f, ok := anyToFloat(v); ok {
				return f, true
			}
		}
	}
	return 0, false
}

func anyToString(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		s := strings.TrimSpace(val)
		if s == "" || s == "0" {
			return "", false
		}
		return s, true
	case float64:
		if val == 0 {
			return "", false
		}
		return strconv.FormatFloat(val, 'f', -1, 64), true
	default:
		return "", false
	}
}

func anyToFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f, err == nil
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// isNotFoundFailMessage 判断 fail message 是否表示“查不到地区”，而非鉴权/配额失败。
//
// 仅匹配保留段、私网、明确的 invalid ip/query 等语义；
// 不含 "invalid key" 这类过宽子串，避免鉴权失败被误判为 NotFound。
func isNotFoundFailMessage(msg string) bool {
	m := strings.ToLower(strings.TrimSpace(msg))
	if m == "" {
		return false
	}
	// 私网 / 保留地址
	if strings.Contains(m, "reserved") ||
		strings.Contains(m, "private") ||
		strings.Contains(m, "bogon") {
		return true
	}
	// 明确的 IP/查询无效（完整短语），排除 invalid key/token/api
	if strings.Contains(m, "invalid ip") ||
		strings.Contains(m, "invalid query") ||
		m == "invalid" {
		return true
	}
	return false
}
