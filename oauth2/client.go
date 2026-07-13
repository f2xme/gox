package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Endpoint 定义标准 OAuth2 端点地址。
type Endpoint struct {
	// AuthURL 授权地址。
	AuthURL string
	// TokenURL 授权码换 token 地址。
	TokenURL string
	// RefreshURL 刷新 token 地址，未设置时复用 TokenURL。
	RefreshURL string
}

// AuthCodeClient 提供标准 OAuth2 授权码流程能力。
type AuthCodeClient struct {
	options Options
}

// NewClient 创建标准 OAuth2 授权码客户端。
func NewClient(opts ...Option) *AuthCodeClient {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &AuthCodeClient{options: options}
}

// AuthCodeURL 生成标准 OAuth2 授权码登录地址。
func (c *AuthCodeClient) AuthCodeURL(state string, opts ...AuthCodeOption) string {
	return BuildAuthCodeURL(AuthCodeURLConfig{
		AuthURL:     c.options.Endpoint.AuthURL,
		ClientID:    c.options.ClientID,
		RedirectURL: c.options.RedirectURL,
	}, state, opts...)
}

// Exchange 使用授权码换取标准 OAuth2 访问令牌。
func (c *AuthCodeClient) Exchange(ctx context.Context, code string) (*Token, error) {
	if code == "" {
		return nil, ErrInvalidCode
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", c.options.ClientID)
	values.Set("client_secret", c.options.ClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", c.options.RedirectURL)
	return c.requestToken(ctx, c.options.Endpoint.TokenURL, values)
}

// RefreshToken 使用刷新令牌续期标准 OAuth2 访问令牌。
func (c *AuthCodeClient) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if refreshToken == "" {
		return nil, ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", c.options.ClientID)
	values.Set("client_secret", c.options.ClientSecret)
	values.Set("refresh_token", refreshToken)
	endpoint := c.options.Endpoint.RefreshURL
	if endpoint == "" {
		endpoint = c.options.Endpoint.TokenURL
	}
	return c.requestToken(ctx, endpoint, values)
}

func (c *AuthCodeClient) requestToken(ctx context.Context, endpoint string, values url.Values) (*Token, error) {
	raw, err := DoPostForm(ctx, c.options.HTTPClient, endpoint, values, "oauth2")
	if err != nil {
		return nil, err
	}
	token, err := DecodeToken(raw)
	if err != nil {
		return nil, fmt.Errorf("oauth2: decode token: %w", err)
	}
	return token, nil
}

// DecodeToken 解析标准 OAuth2 token 响应。
func DecodeToken(raw []byte) (*Token, error) {
	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, &ProviderError{
			Provider: "oauth2",
			Code:     resp.Error,
			Message:  resp.ErrorDescription,
			Raw:      raw,
		}
	}
	if resp.AccessToken == "" {
		return nil, ErrInvalidToken
	}
	return NewToken(TokenInfo{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		Scope:        resp.Scope,
		Raw:          raw,
	}), nil
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
