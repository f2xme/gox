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
	"sync"
	"time"

	"github.com/f2xme/gox/idverify"
)

const (
	codeOK           = "0"
	codeNameMismatch = "222351"
	codeIDInvalid    = "222022"
)

// Verifier 百度 person/idmatch 实现。
type Verifier struct {
	options Options

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

var _ idverify.Verifier = (*Verifier)(nil)

// Provider 返回 baidu。
func (v *Verifier) Provider() string { return idverify.ProviderBaidu }

// Verify 调用百度身份证二要素接口。
func (v *Verifier) Verify(ctx context.Context, req idverify.Request) (idverify.Result, error) {
	start := time.Now()
	if ctx == nil {
		return resultErr(start, fmt.Errorf("%w: context is nil", idverify.ErrInvalidArgument))
	}
	if err := ctx.Err(); err != nil {
		return resultErr(start, err)
	}

	req = req.Normalize()
	if req.Name == "" || req.IDNumber == "" {
		return resultErr(start, fmt.Errorf("%w: name and id number are required", idverify.ErrInvalidArgument))
	}

	token, err := v.getAccessToken(ctx)
	if err != nil {
		return resultErr(start, err)
	}

	body, err := json.Marshal(map[string]string{
		"id_card_number": req.IDNumber,
		"name":           req.Name,
	})
	if err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "marshal", err))
	}

	matchURL, err := url.Parse(v.options.MatchURL)
	if err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "request", err))
	}
	q := matchURL.Query()
	q.Set("access_token", token)
	matchURL.RawQuery = q.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, matchURL.String(), bytes.NewReader(body))
	if err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "request", err))
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := v.options.HTTPClient.Do(httpReq)
	if err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "http", fmt.Errorf("%w: %w", idverify.ErrUnavailable, err)))
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "read", err))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "http",
			fmt.Errorf("%w: status %d: %s", idverify.ErrUnavailable, resp.StatusCode, truncate(string(raw), 200))))
	}

	var parsed matchResp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "decode", err))
	}

	code := normalizeCode(parsed.ErrorCode)
	if code == codeOK || (code == "" && strings.EqualFold(parsed.ErrorMsg, "SUCCESS")) {
		return idverify.Result{
			Matched:  true,
			Provider: idverify.ProviderBaidu,
			Duration: time.Since(start),
		}, nil
	}

	switch code {
	case codeNameMismatch:
		return idverify.Result{
			Matched:      false,
			Provider:     idverify.ProviderBaidu,
			ErrorCode:    idverify.CodeNameMismatch,
			ErrorMessage: "name and id number mismatch",
			ProviderCode: code,
			Duration:     time.Since(start),
		}, nil
	case codeIDInvalid:
		return idverify.Result{
			Matched:      false,
			Provider:     idverify.ProviderBaidu,
			ErrorCode:    idverify.CodeIDInvalid,
			ErrorMessage: "invalid id number",
			ProviderCode: code,
			Duration:     time.Since(start),
		}, nil
	}

	msg := strings.TrimSpace(parsed.ErrorMsg)
	if msg == "" {
		msg = "baidu idmatch error " + code
	}
	return resultErr(start, idverify.Wrap(idverify.ProviderBaidu, "verify",
		fmt.Errorf("%w: error_code=%s error_msg=%s", idverify.ErrUnavailable, code, msg)))
}

type matchResp struct {
	ErrorCode any    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func (v *Verifier) getAccessToken(ctx context.Context) (string, error) {
	v.mu.Lock()
	if v.accessToken != "" && time.Now().Before(v.tokenExpiry) {
		tok := v.accessToken
		v.mu.Unlock()
		return tok, nil
	}
	v.mu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", v.options.APIKey)
	form.Set("client_secret", v.options.SecretKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, v.options.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", idverify.Wrap(idverify.ProviderBaidu, "token", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.options.HTTPClient.Do(httpReq)
	if err != nil {
		return "", idverify.Wrap(idverify.ProviderBaidu, "token", fmt.Errorf("%w: %w", idverify.ErrUnavailable, err))
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", idverify.Wrap(idverify.ProviderBaidu, "token", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", idverify.Wrap(idverify.ProviderBaidu, "token",
			fmt.Errorf("%w: status %d: %s", idverify.ErrUnavailable, resp.StatusCode, truncate(string(raw), 200)))
	}

	var parsed tokenResp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", idverify.Wrap(idverify.ProviderBaidu, "token", err)
	}
	if parsed.AccessToken == "" {
		desc := parsed.ErrorDesc
		if desc == "" {
			desc = parsed.Error
		}
		if desc == "" {
			desc = truncate(string(raw), 200)
		}
		return "", idverify.Wrap(idverify.ProviderBaidu, "token", fmt.Errorf("%w: empty access_token: %s", idverify.ErrUnavailable, desc))
	}

	expiresIn := parsed.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	margin := tokenMargin
	if time.Duration(expiresIn)*time.Second <= margin {
		margin = time.Duration(expiresIn) * time.Second / 2
	}
	expiry := time.Now().Add(time.Duration(expiresIn)*time.Second - margin)

	v.mu.Lock()
	// 并发刷新时可能已有更新 token，仍写入最新成功结果
	v.accessToken = parsed.AccessToken
	v.tokenExpiry = expiry
	tok := v.accessToken
	v.mu.Unlock()
	return tok, nil
}

// resultErr 系统错误：仅填 Provider + Duration，不填业务 ErrorCode。
func resultErr(start time.Time, err error) (idverify.Result, error) {
	return idverify.Result{
		Provider: idverify.ProviderBaidu,
		Duration: time.Since(start),
	}, err
}

func normalizeCode(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(x)
	case float64:
		return fmt.Sprintf("%.0f", x)
	case json.Number:
		return x.String()
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
