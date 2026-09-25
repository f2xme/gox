package x

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

const (
	providerName      = "x"
	defaultUserFields = "id,name,username,profile_image_url,description,location"
)

// ErrMissingCodeVerifier 表示用授权码换票时没有提供 PKCE code_verifier。
var ErrMissingCodeVerifier = errors.New("x: missing pkce code verifier")

// Provider 实现 X 网站应用 OAuth 2.0 授权码登录。
type Provider struct {
	options Options
}

var _ oauth2.Provider = (*Provider)(nil)

// New 创建 X 登录适配器。
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

// Begin 生成带 S256 PKCE 的授权地址，并返回必须保存的 verifier。
func (p *Provider) Begin(state string, opts ...oauth2.AuthCodeOption) (authURL, verifier string, err error) {
	verifier, challenge, err := NewPKCE()
	if err != nil {
		return "", "", err
	}
	all := make([]oauth2.AuthCodeOption, 0, len(opts)+2)
	all = append(all, opts...)
	all = append(all,
		oauth2.WithAuthParam("code_challenge", challenge),
		oauth2.WithAuthParam("code_challenge_method", "S256"),
	)
	return p.AuthCodeURL(state, all...), verifier, nil
}

// AuthCodeURL 生成 X 授权码登录地址。
//
// 默认 scope 为 tweet.read、users.read、offline.access。已传入 code_challenge 但未指定
// code_challenge_method 时，默认使用 S256。推荐用 Begin，避免授权地址和 verifier 对不上。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	applied := oauth2.ApplyAuthCodeOptions(opts...)
	if applied.Extra.Get("code_challenge") != "" && applied.Extra.Get("code_challenge_method") == "" {
		all := make([]oauth2.AuthCodeOption, 0, len(opts)+1)
		all = append(all, opts...)
		all = append(all, oauth2.WithAuthParam("code_challenge_method", "S256"))
		opts = all
	}
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:       p.options.AuthURL,
		ClientID:      p.options.ClientID,
		RedirectURL:   p.options.RedirectURL,
		DefaultScopes: []string{"tweet.read", "users.read", "offline.access"},
	}, state, opts...)
}

// Exchange 使用授权码换取访问令牌。
//
// verifier 通过 WithCodeVerifier 放入 context。没有 verifier 时返回 ErrMissingCodeVerifier。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.ExchangeCode(ctx, code, codeVerifierFrom(ctx))
}

// ExchangeCode 使用授权码和对应的 PKCE verifier 换取访问令牌。
func (p *Provider) ExchangeCode(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	if verifier == "" {
		return nil, ErrMissingCodeVerifier
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", p.options.RedirectURL)
	values.Set("code_verifier", verifier)
	values.Set("client_id", p.options.ClientID)
	return p.requestToken(ctx, p.options.TokenURL, values)
}

// RefreshToken 使用刷新令牌续期访问令牌。
//
// 需要授权时包含 offline.access。X 会轮换 refresh_token，调用方必须保存响应里的新值。
// 响应没有新的刷新令牌时，沿用本次提交的值。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", p.options.ClientID)
	token, err := p.requestToken(ctx, p.refreshEndpoint(), values)
	if err != nil {
		return nil, err
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, nil
}

// UserInfo 使用访问令牌获取当前 X 用户。
//
// ID 和 OpenID 都使用用户编号。用户名、简介和所在地保留在 Raw 中。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	endpoint, err := userEndpoint(p.options.UserURL)
	if err != nil {
		return nil, fmt.Errorf("x: parse user endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("x: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	raw, err := oauth2.DoRequest(p.options.HTTPClient, req, providerName)
	if err != nil {
		return nil, normalizeError(err)
	}

	var resp userEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("x: decode user info: %w", err)
	}
	if resp.Data.ID == "" {
		if code, message, ok := parseXError(raw); ok {
			return nil, &oauth2.ProviderError{
				Provider: providerName,
				Code:     code,
				Message:  message,
				Raw:      raw,
			}
		}
		return nil, oauth2.ErrInvalidToken
	}
	return &oauth2.User{
		Provider:  providerName,
		ID:        resp.Data.ID,
		OpenID:    resp.Data.ID,
		Nickname:  resp.Data.displayName(),
		AvatarURL: resp.Data.ProfileImageURL,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("x: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if p.options.ClientSecret != "" {
		req.SetBasicAuth(p.options.ClientID, p.options.ClientSecret)
	}

	raw, err := oauth2.DoRequest(p.options.HTTPClient, req, providerName)
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
	query.Set("user.fields", defaultUserFields)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func decodeToken(raw []byte) (*oauth2.Token, error) {
	if code, message, ok := parseXError(raw); ok {
		return nil, &oauth2.ProviderError{
			Provider: providerName,
			Code:     code,
			Message:  message,
			Raw:      raw,
		}
	}

	var resp tokenResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("x: decode token: %w", err)
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

func normalizeError(err error) error {
	var pe *oauth2.ProviderError
	if !errors.As(err, &pe) {
		return err
	}
	code, message, ok := parseXError(pe.Raw)
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

func parseXError(raw []byte) (code, message string, ok bool) {
	var body errorBody
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", "", false
	}
	if body.Error != "" {
		return body.Error, body.ErrorDescription, true
	}
	if len(body.Errors) > 0 {
		item := body.Errors[0]
		code = formatCode(item.Code)
		message = item.Message
		if message == "" {
			message = item.Detail
		}
		if message == "" {
			message = item.Title
		}
		if code == "" || code == "0" {
			code = item.Title
		}
		if code == "" && message == "" {
			return "", "", false
		}
		return code, message, true
	}
	if body.Title == "" && body.Detail == "" && body.Status == 0 {
		return "", "", false
	}
	code = strconv.Itoa(body.Status)
	if code == "0" {
		code = body.Title
	}
	message = body.Detail
	if message == "" {
		message = body.Title
	}
	if code == "" && message == "" {
		return "", "", false
	}
	return code, message, true
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
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type errorBody struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Title            string `json:"title"`
	Detail           string `json:"detail"`
	Status           int    `json:"status"`
	Errors           []struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
		Title   string `json:"title"`
		Detail  string `json:"detail"`
	} `json:"errors"`
}

type userEnvelope struct {
	Data userData `json:"data"`
}

type userData struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
}

func (u userData) displayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Username
}
