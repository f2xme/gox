package amap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/f2xme/gox/geo"
)

// Locator 是基于高德 IP 定位 API 的查询实现。
type Locator struct {
	key      string
	endpoint string
	client   *http.Client
}

var _ geo.Locator = (*Locator)(nil)

// amapResponse 高德 IP 定位接口响应。
//
// province/city 在无结果时可能是空字符串或空数组，因此用 json.RawMessage 兼容。
type amapResponse struct {
	Status    string          `json:"status"`
	Info      string          `json:"info"`
	Infocode  string          `json:"infocode"`
	Province  json.RawMessage `json:"province"`
	City      json.RawMessage `json:"city"`
	Adcode    json.RawMessage `json:"adcode"`
	Rectangle string          `json:"rectangle"`
}

// Lookup 通过高德 IP 定位接口查询地区信息。
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
	query.Set("ip", normalized)
	query.Set("key", l.key)
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "create request failed", err, normalized)
	}
	req.Header.Set("Accept", "application/json")

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

	var raw amapResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "parse response failed", err, normalized)
	}

	if raw.Status != "1" {
		msg := strings.TrimSpace(raw.Info)
		if msg == "" {
			msg = "amap lookup failed"
		}
		// Key 无效、权限不足、配额耗尽等均属上游/配置故障，统一 Unavailable，
		// 避免调用方按 InvalidArgument 误做“修正入参后重试”。
		return nil, geo.NewError(geo.ErrCodeUnavailable, msg, normalized)
	}

	province := decodeAmapString(raw.Province)
	city := decodeAmapString(raw.City)
	adcode := decodeAmapString(raw.Adcode)

	loc := &geo.Location{
		IP:          normalized,
		Country:     "中国",
		CountryCode: "CN",
		Province:    province,
		City:        city,
	}
	if adcode != "" || raw.Rectangle != "" {
		loc.Extra = make(map[string]string)
		if adcode != "" {
			loc.Extra["adcode"] = adcode
		}
		if raw.Rectangle != "" {
			loc.Extra["rectangle"] = raw.Rectangle
		}
	}

	// 局域网/无效 IP 时高德常返回 status=1 但 province/city 为空
	if loc.Province == "" && loc.City == "" {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	return loc, nil
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
	// 非字符串时忽略
	return ""
}
