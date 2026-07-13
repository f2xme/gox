package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/f2xme/gox/oauth2"
)

const providerName = "wechat"

// Provider 实现微信开放平台网站应用登录。
type Provider struct {
	options Options
}

// New 创建微信登录适配器。
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

// AuthCodeURL 生成微信扫码登录授权地址。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:        p.options.AuthURL,
		ClientID:       p.options.ClientID,
		ClientIDParam:  "appid",
		RedirectURL:    p.options.RedirectURL,
		ScopeSeparator: ",",
		DefaultScopes:  []string{"snsapi_login"},
		Fragment:       "wechat_redirect",
	}, state, opts...)
}

// Exchange 使用授权码换取微信访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("appid", p.options.ClientID)
	values.Set("secret", p.options.ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	return p.requestToken(ctx, p.options.TokenURL, values)
}

// RefreshToken 使用刷新令牌续期微信访问令牌。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("appid", p.options.ClientID)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	return p.requestToken(ctx, p.options.RefreshURL, values)
}

// UserInfo 使用访问令牌获取微信用户信息。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" || token.OpenID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	values := url.Values{}
	values.Set("access_token", token.AccessToken)
	values.Set("openid", token.OpenID)
	values.Set("lang", "zh_CN")

	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, p.options.UserURL, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp userResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("wechat: decode user info: %w", err)
	}
	if resp.ErrCode != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(resp.ErrCode),
			Message:  resp.ErrMsg,
			Raw:      raw,
		}
	}

	return &oauth2.User{
		Provider:  providerName,
		ID:        resp.OpenID,
		OpenID:    resp.OpenID,
		UnionID:   resp.UnionID,
		Nickname:  resp.Nickname,
		AvatarURL: resp.HeadImgURL,
		Gender:    wechatGender(resp.Sex),
		Country:   resp.Country,
		Province:  resp.Province,
		City:      resp.City,
		Raw:       raw,
	}, nil
}

func (p *Provider) requestToken(ctx context.Context, endpoint string, values url.Values) (*oauth2.Token, error) {
	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, endpoint, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("wechat: decode token: %w", err)
	}
	if resp.ErrCode != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(resp.ErrCode),
			Message:  resp.ErrMsg,
			Raw:      raw,
		}
	}
	if resp.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}

	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		OpenID:       resp.OpenID,
		UnionID:      resp.UnionID,
		Scope:        resp.Scope,
		Raw:          raw,
	}), nil
}

func wechatGender(sex int) string {
	switch sex {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return "unknown"
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

type userResponse struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}
