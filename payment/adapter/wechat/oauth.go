package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/f2xme/gox/payment"
)

type oauthHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// OAuthURL 创建微信网页授权地址，scope 固定为 snsapi_base。
func (w *WechatPay) OAuthURL(redirectURL, state string) (string, error) {
	if w.config.OAuthAppSecret == "" {
		return "", fmt.Errorf("%w: wechat OAuth app secret cannot be empty", payment.ErrInvalidConfig)
	}
	u, err := url.Parse(redirectURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", fmt.Errorf("%w: OAuth redirect URL must be an absolute HTTPS URL", payment.ErrInvalidRequest)
	}
	if state == "" {
		return "", fmt.Errorf("%w: OAuth state cannot be empty", payment.ErrInvalidRequest)
	}
	q := url.Values{"appid": {w.config.AppID}, "redirect_uri": {redirectURL}, "response_type": {"code"}, "scope": {"snsapi_base"}, "state": {state}}
	return w.oauthAuthURL + "?" + q.Encode() + "#wechat_redirect", nil
}

// ExchangeOAuthCode 使用网页授权 code 换取 openid。
func (w *WechatPay) ExchangeOAuthCode(ctx context.Context, code string) (string, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return "", err
	}
	if w.config.OAuthAppSecret == "" {
		return "", fmt.Errorf("%w: wechat OAuth app secret cannot be empty", payment.ErrInvalidConfig)
	}
	if code == "" {
		return "", fmt.Errorf("%w: OAuth code cannot be empty", payment.ErrInvalidRequest)
	}
	u, _ := url.Parse(w.oauthTokenURL)
	q := u.Query()
	q.Set("appid", w.config.AppID)
	q.Set("secret", w.config.OAuthAppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: create OAuth request", payment.ErrInvalidRequest)
	}
	resp, err := w.oauthClient.Do(req)
	if err != nil {
		return "", providerError("oauth_exchange", err)
	}
	defer resp.Body.Close()
	var result struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", providerError("oauth_exchange", err)
	}
	if resp.StatusCode != http.StatusOK || result.ErrCode != 0 || result.OpenID == "" {
		return "", &payment.ProviderError{Provider: payment.ProviderWechat, Operation: "oauth_exchange", Code: strconv.Itoa(result.ErrCode), Message: result.ErrMsg, Err: payment.ErrGateway}
	}
	return result.OpenID, nil
}
