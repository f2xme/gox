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
	body := trimCallback(raw)

	var resp userResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     "decode_userinfo",
			Message:  err.Error(),
			Raw:      raw,
		}
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
	// QQ 可能返回 JSON、form，或 JSONP（callback(...)）；Content-Type 偶发 text/html。
	body := trimCallback(raw)

	accessToken, refreshToken, tokenType, scope, errCode, errDesc, expiresIn, parseOK := parseTokenBody(body)
	if !parseOK {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     "decode_token",
			Message:  "qq token response unreadable",
			Raw:      raw,
		}
	}
	if errCode != "" && errCode != "0" {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     errCode,
			Message:  firstNonEmpty(errDesc, "qq token error"),
			Raw:      raw,
		}
	}
	if accessToken == "" {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     "empty_access_token",
			Message:  firstNonEmpty(errDesc, "qq access_token empty"),
			Raw:      raw,
		}
	}

	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken:  accessToken,
		TokenType:    tokenType,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Scope:        scope,
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
	body := trimCallback(raw)

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     "decode_openid",
			Message:  err.Error(),
			Raw:      raw,
		}
	}
	if code := anyString(m["error"]); code != "" && code != "0" {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  firstNonEmpty(anyString(m["error_description"]), "qq openid error"),
			Raw:      raw,
		}
	}
	openID := anyString(m["openid"])
	if openID == "" {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     "empty_openid",
			Message:  "qq openid empty",
			Raw:      raw,
		}
	}
	token.OpenID = openID
	token.UnionID = anyString(m["unionid"])
	return token, nil
}

// parseTokenBody 兼容 JSON / form 的 token 响应；error / expires_in 支持数字与字符串。
func parseTokenBody(body []byte) (accessToken, refreshToken, tokenType, scope, errCode, errDesc string, expiresIn int64, ok bool) {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return "", "", "", "", "", "", 0, false
	}

	if strings.HasPrefix(text, "{") {
		var m map[string]any
		if err := json.Unmarshal([]byte(text), &m); err == nil {
			return anyString(m["access_token"]),
				anyString(m["refresh_token"]),
				anyString(m["token_type"]),
				anyString(m["scope"]),
				anyString(m["error"]),
				anyString(m["error_description"]),
				anyInt64(m["expires_in"]),
				true
		}
	}

	if query, err := url.ParseQuery(text); err == nil && (query.Get("access_token") != "" || query.Get("error") != "") {
		expiresIn, _ = strconv.ParseInt(query.Get("expires_in"), 10, 64)
		return query.Get("access_token"),
			query.Get("refresh_token"),
			query.Get("token_type"),
			query.Get("scope"),
			query.Get("error"),
			query.Get("error_description"),
			expiresIn,
			true
	}

	return "", "", "", "", "", "", 0, false
}

func anyString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}

func anyInt64(v any) int64 {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return 0
	}
}

// trimCallback 去掉 QQ 常见 JSONP 外壳 callback(...)。
func trimCallback(raw []byte) []byte {
	text := strings.TrimSpace(string(raw))
	if strings.HasPrefix(text, "callback(") {
		text = strings.TrimPrefix(text, "callback(")
		text = strings.TrimSpace(text)
		text = strings.TrimSuffix(text, ";")
		text = strings.TrimSpace(text)
		text = strings.TrimSuffix(text, ")")
		text = strings.TrimSpace(text)
	}
	return []byte(text)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
