package facebook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/f2xme/gox/oauth2"
)

const (
	providerName      = "facebook"
	defaultUserFields = "id,name,email,picture.type(large)"
)

// Provider 实现 Facebook 网站应用授权码登录。
type Provider struct {
	options Options
}

var _ oauth2.Provider = (*Provider)(nil)

// New 创建 Facebook 登录适配器。
func New(opts ...Option) *Provider {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &Provider{options: options}
}

// Name 返回服务提供商名称。
func (p *Provider) Name() string {
	return providerName
}

// AuthCodeURL 生成 Facebook 授权码登录地址。
//
// 默认 scope 为 public_profile、email，多项 scope 使用逗号连接。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:        p.options.AuthURL,
		ClientID:       p.options.ClientID,
		RedirectURL:    p.options.RedirectURL,
		ScopeSeparator: ",",
		DefaultScopes:  []string{"public_profile", "email"},
	}, state, opts...)
}

// Exchange 使用授权码换取 Facebook 短期用户访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("redirect_uri", p.options.RedirectURL)
	values.Set("code", code)
	return p.requestToken(ctx, p.options.TokenURL, values)
}

// RefreshToken 把短期用户访问令牌换成长期用户访问令牌。
//
// Facebook 不签发 refresh_token。这里的参数是仍有效的短期 access_token，
// 不是 OAuth 刷新令牌。已过期的令牌无法续期，需要重新走授权。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "fb_exchange_token")
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("fb_exchange_token", refreshToken)
	return p.requestToken(ctx, p.refreshEndpoint(), values)
}

// UserInfo 使用访问令牌获取 Facebook 用户信息。
//
// ID 和 OpenID 都使用应用范围内的用户编号。Facebook 没有 unionid。
// 邮箱保留在 Raw 中；用户拒绝 email 权限时该字段可能不存在。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	endpoint, err := userEndpoint(p.options.UserURL)
	if err != nil {
		return nil, fmt.Errorf("facebook: parse user endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("facebook: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	raw, err := oauth2.DoRequest(p.options.HTTPClient, req, providerName)
	if err != nil {
		return nil, normalizeError(err)
	}
	if code, message, ok := parseFacebookError(raw); ok {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  message,
			Raw:      raw,
		}
	}

	var resp userResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("facebook: decode user info: %w", err)
	}
	if resp.ID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return &oauth2.User{
		Provider:  providerName,
		ID:        resp.ID,
		OpenID:    resp.ID,
		Nickname:  resp.Name,
		AvatarURL: resp.Picture.Data.URL,
		Raw:       raw,
	}, nil
}

func (p *Provider) refreshEndpoint() string {
	if p.options.RefreshURL != "" {
		return p.options.RefreshURL
	}
	return p.options.TokenURL
}

func (p *Provider) requestToken(ctx context.Context, endpoint string, values url.Values) (*oauth2.Token, error) {
	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, endpoint, values, providerName)
	if err != nil {
		return nil, normalizeError(err)
	}
	return decodeToken(raw)
}

func userEndpoint(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("fields", defaultUserFields)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func decodeToken(raw []byte) (*oauth2.Token, error) {
	if code, message, ok := parseFacebookError(raw); ok {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  message,
			Raw:      raw,
		}
	}

	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("facebook: decode token: %w", err)
	}
	if resp.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		ExpiresIn:   resp.ExpiresIn,
		Raw:         raw,
	}), nil
}

// normalizeError 把 Facebook 非 2xx 响应体里的 error 对象提取成平台错误码。
func normalizeError(err error) error {
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) {
		return err
	}
	code, message, ok := parseFacebookError(pe.Raw)
	if !ok {
		return err
	}
	return &oauth2.ProviderError{
		Provider:   providerName,
		Code:       code,
		Message:    message,
		Raw:        pe.Raw,
		StatusCode: pe.StatusCode,
	}
}

// parseFacebookError 解析 Graph API 的 error 对象。
// 成功响应没有 error 字段时返回 ok=false。
func parseFacebookError(raw []byte) (code, message string, ok bool) {
	var body struct {
		Error            json.RawMessage `json:"error"`
		ErrorDescription string          `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &body); err != nil || len(body.Error) == 0 || string(body.Error) == "null" {
		return "", "", false
	}
	switch body.Error[0] {
	case '"':
		if err := json.Unmarshal(body.Error, &code); err != nil || code == "" {
			return "", "", false
		}
		return code, body.ErrorDescription, true
	case '{':
		var obj struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    any    `json:"code"`
		}
		if err := json.Unmarshal(body.Error, &obj); err != nil {
			return "", "", false
		}
		code = formatCode(obj.Code)
		if code == "" || code == "0" {
			code = obj.Type
		}
		if code == "" && obj.Message == "" {
			return "", "", false
		}
		return code, obj.Message, true
	default:
		return "", "", false
	}
}

func formatCode(v any) string {
	switch n := v.(type) {
	case nil:
		return ""
	case string:
		return n
	case float64:
		if n == 0 {
			return "0"
		}
		return strconv.FormatInt(int64(n), 10)
	default:
		return fmt.Sprint(n)
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type userResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}
