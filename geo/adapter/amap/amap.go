package amap

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec // 高德 Web 服务协议要求使用 MD5 生成 sig。
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/f2xme/gox/geo"
)

// Locator 是基于高德 IP 定位 API 的查询实现。
type Locator struct {
	key        string
	privateKey string
	endpoint   string
	client     *http.Client
}

var _ geo.Locator = (*Locator)(nil)

// amapResponse 高德 IP 定位接口响应。
//
// province/city 在无结果时可能是空字符串或空数组，因此用 json.RawMessage 兼容。
type amapResponse struct {
	Status    json.RawMessage `json:"status"`
	Info      string          `json:"info"`
	Infocode  string          `json:"infocode"`
	Province  json.RawMessage `json:"province"`
	City      json.RawMessage `json:"city"`
	Adcode    json.RawMessage `json:"adcode"`
	Rectangle string          `json:"rectangle"`
}

// Lookup 通过高德 IP 定位接口查询地区信息。
func (l *Locator) Lookup(ctx context.Context, ip string) (*geo.Location, error) {
	if err := validateLookupContext(ctx, ip); err != nil {
		return nil, err
	}

	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return nil, err
	}
	if net.ParseIP(normalized).To4() == nil {
		return nil, geo.NewError(geo.ErrCodeInvalidIP, "amap IP location only supports IPv4", normalized)
	}
	return l.lookup(ctx, normalized)
}

// LookupCurrent 省略 ip 参数，通过高德定位当前 HTTP 请求来源 IP。
//
// 高德响应不包含来源 IP，因此成功结果的 Location.IP 为空。
func (l *Locator) LookupCurrent(ctx context.Context) (*geo.Location, error) {
	if err := validateLookupContext(ctx, ""); err != nil {
		return nil, err
	}
	return l.lookup(ctx, "")
}

func (l *Locator) lookup(ctx context.Context, ip string) (*geo.Location, error) {
	reqURL, err := url.Parse(l.endpoint)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInvalidArgument, "invalid endpoint", err, ip)
	}
	query := reqURL.Query()
	if ip == "" {
		query.Del("ip")
	} else {
		query.Set("ip", ip)
	}
	query.Set("key", l.key)
	query.Set("output", "json")
	if l.privateKey == "" {
		query.Del("sig")
	} else {
		query.Set("sig", signAmapQuery(query, l.privateKey))
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "create request failed", err, ip)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeUnavailable, "http request failed", err, ip)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "read response failed", err, ip)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, geo.NewError(
			geo.ErrCodeUnavailable,
			fmt.Sprintf("unexpected status code %d", resp.StatusCode),
			ip,
		)
	}

	var raw amapResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "parse response failed", err, ip)
	}

	status := decodeAmapString(raw.Status)
	infocode := strings.TrimSpace(raw.Infocode)
	if status != "1" || (infocode != "" && infocode != "10000") {
		msg := strings.TrimSpace(raw.Info)
		if msg == "" {
			msg = "amap lookup failed"
		}
		if infocode != "" {
			msg += " (infocode=" + infocode + ")"
		}
		// Key 无效、权限不足、配额耗尽等均属上游/配置故障，统一 Unavailable，
		// 避免调用方按 InvalidArgument 误做“修正入参后重试”。
		return nil, geo.NewError(geo.ErrCodeUnavailable, msg, ip)
	}

	province := decodeAmapString(raw.Province)
	city := decodeAmapString(raw.City)
	adcode := decodeAmapString(raw.Adcode)

	loc := &geo.Location{
		IP:          ip,
		Country:     "中国",
		CountryCode: "CN",
		Province:    province,
		City:        city,
	}
	rectangle := strings.TrimSpace(raw.Rectangle)
	if adcode != "" || rectangle != "" {
		loc.Extra = make(map[string]string)
		if adcode != "" {
			loc.Extra["adcode"] = adcode
		}
		if rectangle != "" {
			loc.Extra["rectangle"] = rectangle
		}
	}

	// 局域网返回 province="局域网"；非法或国外 IP 返回空省市。
	if (loc.Province == "" && loc.City == "") ||
		loc.Province == "局域网" || loc.City == "局域网" {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", ip)
	}
	return loc, nil
}

func validateLookupContext(ctx context.Context, ip string) error {
	if ctx == nil {
		return geo.NewError(geo.ErrCodeInvalidArgument, "context cannot be nil", ip)
	}
	if err := ctx.Err(); err != nil {
		return geo.WrapError(geo.ErrCodeInternal, "context error", err, ip)
	}
	return nil
}

// signAmapQuery 按高德协议生成请求签名：参数名升序拼接后追加私钥并计算 MD5。
func signAmapQuery(values url.Values, privateKey string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "sig" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	var content strings.Builder
	for _, key := range keys {
		for _, value := range values[key] {
			if content.Len() > 0 {
				content.WriteByte('&')
			}
			content.WriteString(key)
			content.WriteByte('=')
			content.WriteString(value)
		}
	}
	content.WriteString(privateKey)
	sum := md5.Sum([]byte(content.String()))
	return fmt.Sprintf("%x", sum)
}

// decodeAmapString 解析高德字段：可能是 JSON 字符串或空数组 []。
func decodeAmapString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.TrimSpace(s)
		if s == "" || s == "[]" {
			return ""
		}
		return s
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err == nil {
		return number.String()
	}
	// 非字符串时忽略
	return ""
}
