package oauth2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const maxResponseBodyBytes = int64(1 << 20)

// DoGet 发送带查询参数的 GET 请求并返回响应体。
func DoGet(ctx context.Context, client *http.Client, endpoint string, values url.Values, provider string) ([]byte, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("%s: parse endpoint: %w", provider, err)
	}
	query := u.Query()
	for key, vals := range values {
		for _, val := range vals {
			query.Add(key, val)
		}
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", provider, err)
	}
	return DoRequest(client, req, provider)
}

// DoPostForm 发送 application/x-www-form-urlencoded 请求并返回响应体。
func DoPostForm(ctx context.Context, client *http.Client, endpoint string, values url.Values, provider string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", provider, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return DoRequest(client, req, provider)
}

// DoRequest 执行 HTTP 请求并把非 2xx 响应转换为 ProviderError。
func DoRequest(client *http.Client, req *http.Request, provider string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", provider, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", provider, err)
	}
	if int64(len(raw)) > maxResponseBodyBytes {
		return nil, fmt.Errorf("%s: response exceeds %d bytes", provider, maxResponseBodyBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &ProviderError{
			Provider:   provider,
			Code:       strconv.Itoa(resp.StatusCode),
			Message:    string(raw),
			Raw:        raw,
			StatusCode: resp.StatusCode,
		}
	}
	return raw, nil
}
