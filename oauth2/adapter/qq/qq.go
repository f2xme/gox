package qq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/f2xme/gox/oauth2"
)

const providerName = "qq"

// Provider 实现 QQ 互联网站应用登录。
type Provider struct {
	options Options
}

// New 创建 QQ 登录适配器。
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

// AuthCodeURL 生成 QQ 登录授权地址。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:        p.options.AuthURL,
		ClientID:       p.options.ClientID,
		RedirectURL:    p.options.RedirectURL,
		ScopeSeparator: ",",
		DefaultScopes:  []string{"get_user_info"},
	}, state, opts...)
}

// Exchange 使用授权码换取 QQ 访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", p.options.RedirectURL)
	values.Set("fmt", "json")

	token, err := p.requestToken(ctx, values)
	if err != nil {
		return nil, err
	}
	return p.fillOpenID(ctx, token)
}

// RefreshToken 使用刷新令牌续期 QQ 访问令牌。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", p.options.ClientID)
	values.Set("client_secret", p.options.ClientSecret)
	values.Set("refresh_token", refreshToken)
	values.Set("fmt", "json")

	token, err := p.requestToken(ctx, values)
	if err != nil {
		return nil, err
	}
	return p.fillOpenID(ctx, token)
}

// UserInfo 使用访问令牌获取 QQ 用户信息。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" || token.OpenID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	values := url.Values{}
	values.Set("access_token", token.AccessToken)
	values.Set("oauth_consumer_key", p.options.ClientID)
	values.Set("openid", token.OpenID)

	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, p.options.UserURL, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp userResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("qq: decode user info: %w", err)
	}
	if resp.Ret != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(resp.Ret),
			Message:  resp.Msg,
			Raw:      raw,
		}
	}

	return &oauth2.User{
		Provider:  providerName,
		ID:        token.OpenID,
		OpenID:    token.OpenID,
		UnionID:   token.UnionID,
		Nickname:  resp.Nickname,
		AvatarURL: firstNonEmpty(resp.FigureURLQQ2, resp.FigureURLQQ1, resp.FigureURL2, resp.FigureURL1, resp.FigureURL),
		Gender:    resp.Gender,
		Province:  resp.Province,
		City:      resp.City,
		Raw:       raw,
	}, nil
}

func (p *Provider) requestToken(ctx context.Context, values url.Values) (*oauth2.Token, error) {
	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, p.options.TokenURL, values, providerName)
	if err != nil {
		return nil, err
	}

	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		query, parseErr := url.ParseQuery(string(raw))
		if parseErr != nil {
			return nil, fmt.Errorf("qq: decode token: %w", err)
		}
		resp.AccessToken = query.Get("access_token")
		resp.RefreshToken = query.Get("refresh_token")
		resp.Scope = query.Get("scope")
		resp.ExpiresIn, _ = strconv.ParseInt(query.Get("expires_in"), 10, 64)
	}
	if resp.Error != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(resp.Error),
			Message:  resp.ErrorDescription,
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
		Scope:        resp.Scope,
		Raw:          raw,
	}), nil
}

func (p *Provider) fillOpenID(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	values := url.Values{}
	values.Set("access_token", token.AccessToken)
	values.Set("fmt", "json")

	raw, err := oauth2.DoGet(ctx, p.options.HTTPClient, p.options.OpenIDURL, values, providerName)
	if err != nil {
		return nil, err
	}
	var resp openIDResponse
	if err := json.Unmarshal(trimCallback(raw), &resp); err != nil {
		return nil, fmt.Errorf("qq: decode openid: %w", err)
	}
	if resp.Error != 0 {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     strconv.Itoa(resp.Error),
			Message:  resp.ErrorDescription,
			Raw:      raw,
		}
	}
	if resp.OpenID == "" {
		return nil, oauth2.ErrInvalidToken
	}
	token.OpenID = resp.OpenID
	token.UnionID = resp.UnionID
	return token, nil
}

func trimCallback(raw []byte) []byte {
	text := strings.TrimSpace(string(raw))
	if strings.HasPrefix(text, "callback(") && strings.HasSuffix(text, ");") {
		text = strings.TrimSuffix(strings.TrimPrefix(text, "callback("), ");")
	}
	return []byte(strings.TrimSpace(text))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	Error            int    `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type openIDResponse struct {
	ClientID         string `json:"client_id"`
	OpenID           string `json:"openid"`
	UnionID          string `json:"unionid"`
	Error            int    `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type userResponse struct {
	Ret          int    `json:"ret"`
	Msg          string `json:"msg"`
	Nickname     string `json:"nickname"`
	FigureURL    string `json:"figureurl"`
	FigureURL1   string `json:"figureurl_1"`
	FigureURL2   string `json:"figureurl_2"`
	FigureURLQQ1 string `json:"figureurl_qq_1"`
	FigureURLQQ2 string `json:"figureurl_qq_2"`
	Gender       string `json:"gender"`
	Province     string `json:"province"`
	City         string `json:"city"`
}
