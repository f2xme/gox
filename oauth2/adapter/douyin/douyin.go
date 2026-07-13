package douyin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/f2xme/gox/oauth2"
)

const providerName = "douyin"

// Provider 实现抖音开放平台网站应用登录。
type Provider struct {
	options Options
}

// New 创建抖音登录适配器。
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

// AuthCodeURL 生成抖音登录授权地址。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:        p.options.AuthURL,
		ClientID:       p.options.ClientID,
		ClientIDParam:  "client_key",
		RedirectURL:    p.options.RedirectURL,
		ScopeSeparator: ",",
		DefaultScopes:  []string{"user_info"},
	}, state, opts...)
}

// Exchange 使用授权码换取抖音访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("client_key", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	return p.requestToken(ctx, p.options.TokenURL, values)
}

// RefreshToken 使用刷新令牌续期抖音访问令牌。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("client_key", p.options.ClientID)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	return p.requestToken(ctx, p.options.RefreshURL, values)
}

// UserInfo 使用访问令牌获取抖音用户信息。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" || token.OpenID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	values := url.Values{}
	values.Set("access_token", token.AccessToken)
	values.Set("open_id", token.OpenID)

	raw, err := oauth2.DoPostForm(ctx, p.options.HTTPClient, p.options.UserURL, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp userEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("douyin: decode user info: %w", err)
	}
	if code, message := resp.errorInfo(); code != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(code),
			Message:  message,
			Raw:      raw,
		}
	}

	data := resp.Data
	if data.OpenID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return &oauth2.User{
		Provider:  providerName,
		ID:        data.OpenID,
		OpenID:    data.OpenID,
		UnionID:   data.UnionID,
		Nickname:  data.Nickname,
		AvatarURL: firstNonEmpty(data.AvatarLarger, data.Avatar),
		Gender:    data.Gender,
		Country:   data.Country,
		Province:  data.Province,
		City:      data.City,
		Raw:       raw,
	}, nil
}

func (p *Provider) requestToken(ctx context.Context, endpoint string, values url.Values) (*oauth2.Token, error) {
	raw, err := oauth2.DoPostForm(ctx, p.options.HTTPClient, endpoint, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp tokenEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("douyin: decode token: %w", err)
	}
	if code, message := resp.errorInfo(); code != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(code),
			Message:  message,
			Raw:      raw,
		}
	}

	data := resp.Data
	if data.AccessToken == "" && resp.AccessToken != "" {
		data.AccessToken = resp.AccessToken
		data.TokenType = resp.TokenType
		data.ExpiresIn = resp.ExpiresIn
		data.RefreshToken = resp.RefreshToken
		data.OpenID = resp.OpenID
		data.Scope = resp.Scope
	}
	if data.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}

	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken:  data.AccessToken,
		TokenType:    data.TokenType,
		RefreshToken: data.RefreshToken,
		ExpiresIn:    int64(data.ExpiresIn),
		OpenID:       data.OpenID,
		UnionID:      data.UnionID,
		Scope:        data.Scope,
		Raw:          raw,
	}), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type tokenEnvelope struct {
	Message      string        `json:"message"`
	ErrorCode    flexibleInt64 `json:"error_code"`
	Description  string        `json:"description"`
	AccessToken  string        `json:"access_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    flexibleInt64 `json:"expires_in"`
	RefreshToken string        `json:"refresh_token"`
	OpenID       string        `json:"open_id"`
	Scope        string        `json:"scope"`
	Data         tokenData     `json:"data"`
}

func (e tokenEnvelope) errorInfo() (int, string) {
	if e.ErrorCode != 0 {
		return int(e.ErrorCode), firstNonEmpty(e.Description, e.Message)
	}
	if e.Data.ErrorCode != 0 {
		return int(e.Data.ErrorCode), firstNonEmpty(e.Data.Description, e.Data.Message)
	}
	return 0, ""
}

type tokenData struct {
	AccessToken  string        `json:"access_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    flexibleInt64 `json:"expires_in"`
	RefreshToken string        `json:"refresh_token"`
	OpenID       string        `json:"open_id"`
	UnionID      string        `json:"union_id"`
	Scope        string        `json:"scope"`
	ErrorCode    flexibleInt64 `json:"error_code"`
	Description  string        `json:"description"`
	Message      string        `json:"message"`
}

type userEnvelope struct {
	Message     string        `json:"message"`
	ErrorCode   flexibleInt64 `json:"error_code"`
	Description string        `json:"description"`
	ErrNo       flexibleInt64 `json:"err_no"`
	ErrMsg      string        `json:"err_msg"`
	Data        userData      `json:"data"`
}

func (e userEnvelope) errorInfo() (int, string) {
	if e.ErrNo != 0 {
		return int(e.ErrNo), firstNonEmpty(e.ErrMsg, e.Description, e.Message)
	}
	if e.ErrorCode != 0 {
		return int(e.ErrorCode), firstNonEmpty(e.Description, e.Message)
	}
	if e.Data.ErrorCode != 0 {
		return int(e.Data.ErrorCode), firstNonEmpty(e.Data.Description, e.Data.Message)
	}
	return 0, ""
}

type userData struct {
	OpenID       string        `json:"open_id"`
	UnionID      string        `json:"union_id"`
	Nickname     string        `json:"nickname"`
	Avatar       string        `json:"avatar"`
	AvatarLarger string        `json:"avatar_larger"`
	Gender       string        `json:"gender"`
	Country      string        `json:"country"`
	Province     string        `json:"province"`
	City         string        `json:"city"`
	ErrorCode    flexibleInt64 `json:"error_code"`
	Description  string        `json:"description"`
	Message      string        `json:"message"`
}

type flexibleInt64 int64

func (v *flexibleInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = 0
		return nil
	}
	text := string(data)
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	}
	if text == "" {
		*v = 0
		return nil
	}
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return fmt.Errorf("douyin: parse integer %q: %w", text, err)
	}
	*v = flexibleInt64(value)
	return nil
}
