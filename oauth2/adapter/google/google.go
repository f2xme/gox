package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f2xme/gox/oauth2"
)

const providerName = "google"

// Provider 实现 Google 网站应用 OpenID Connect 登录。
type Provider struct {
	options Options
}

var _ oauth2.Provider = (*Provider)(nil)

// New 创建 Google 登录适配器。
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

// AuthCodeURL 生成 Google 授权码登录地址。
//
// 默认 scope 为 openid、email、profile，并带上 access_type=offline，以便 Google 签发 refresh_token。
// 调用方可以通过 oauth2.WithScopes 或 oauth2.WithAuthParam 覆盖这两项。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	opts = append([]oauth2.AuthCodeOption{
		oauth2.WithAuthParam("access_type", "offline"),
	}, opts...)
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:       p.options.AuthURL,
		ClientID:      p.options.ClientID,
		RedirectURL:   p.options.RedirectURL,
		DefaultScopes: []string{"openid", "email", "profile"},
	}, state, opts...)
}

// Exchange 使用授权码换取 Google 访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("redirect_uri", p.options.RedirectURL)
	return p.requestToken(ctx, p.options.TokenURL, values)
}

// RefreshToken 使用刷新令牌续期 Google 访问令牌。
//
// Google 刷新响应通常不再返回 refresh_token。响应里没有新的刷新令牌时，沿用本次提交的值。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	token, err := p.requestToken(ctx, p.refreshEndpoint(), values)
	if err != nil {
		return nil, err
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, nil
}

// UserInfo 使用访问令牌获取 Google 用户信息。
//
// 用户唯一标识使用 OpenID Connect 的 sub。Google 没有 openid 与 unionid 之分，
// 因此 ID 和 OpenID 都填 sub；旧版 userinfo 只返回 id 时，用 id 兜底。
// 邮箱、email_verified、locale 保留在 Raw 中。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.options.UserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("google: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	raw, err := oauth2.DoRequest(p.options.HTTPClient, req, providerName)
	if err != nil {
		return nil, normalizeError(err)
	}
	if code, message, ok := parseGoogleError(raw); ok {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  message,
			Raw:      raw,
		}
	}

	var resp userResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("google: decode user info: %w", err)
	}
	subject := resp.subject()
	if subject == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return &oauth2.User{
		Provider:  providerName,
		ID:        subject,
		OpenID:    subject,
		Nickname:  resp.displayName(),
		AvatarURL: resp.Picture,
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
	raw, err := oauth2.DoPostForm(ctx, p.options.HTTPClient, endpoint, values, providerName)
	if err != nil {
		return nil, normalizeError(err)
	}
	return decodeToken(raw)
}

func decodeToken(raw []byte) (*oauth2.Token, error) {
	if code, message, ok := parseGoogleError(raw); ok {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  message,
			Raw:      raw,
		}
	}

	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("google: decode token: %w", err)
	}
	if resp.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		Scope:        resp.Scope,
		Raw:          raw,
	}), nil
}

// normalizeError 把 Google 非 2xx 响应体里的 error 字段提取成平台错误码。
func normalizeError(err error) error {
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) {
		return err
	}
	code, message, ok := parseGoogleError(pe.Raw)
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

// parseGoogleError 解析 OAuth 字符串错误或 Google API 对象错误。
// 成功响应没有 error 字段时返回 ok=false。
func parseGoogleError(raw []byte) (code, message string, ok bool) {
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
			Code    any    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal(body.Error, &obj); err != nil {
			return "", "", false
		}
		code = obj.Status
		if code == "" {
			code = formatCode(obj.Code)
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
		return strconv.FormatInt(int64(n), 10)
	default:
		return fmt.Sprint(n)
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type userResponse struct {
	Sub        string `json:"sub"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
}

func (r userResponse) subject() string {
	if r.Sub != "" {
		return r.Sub
	}
	return r.ID
}

func (r userResponse) displayName() string {
	if r.Name != "" {
		return r.Name
	}
	return strings.TrimSpace(r.GivenName + " " + r.FamilyName)
}
