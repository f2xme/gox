package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/f2xme/gox/oauth2"
)

const providerName = "alipay"

// Provider 实现支付宝开放平台网页授权登录。
type Provider struct {
	options Options
}

// New 创建支付宝登录适配器。
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

// AuthCodeURL 生成支付宝授权登录地址。
func (p *Provider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2.BuildAuthCodeURL(oauth2.AuthCodeURLConfig{
		AuthURL:        p.options.AuthURL,
		ClientID:       p.options.ClientID,
		ClientIDParam:  "app_id",
		RedirectURL:    p.options.RedirectURL,
		ScopeSeparator: ",",
		DefaultScopes:  []string{"auth_user"},
	}, state, opts...)
}

// Exchange 使用授权码换取支付宝访问令牌。
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	return p.requestToken(ctx, values)
}

// RefreshToken 使用刷新令牌续期支付宝访问令牌。
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrMissingRefreshToken
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	return p.requestToken(ctx, values)
}

// UserInfo 使用访问令牌获取支付宝用户信息。
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.User, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}

	values := url.Values{}
	values.Set("auth_token", token.AccessToken)
	raw, err := p.call(ctx, "alipay.user.info.share", values)
	if err != nil {
		return nil, err
	}

	var resp userEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("alipay: decode user info: %w", err)
	}
	if err := p.verifyResponse(raw, "alipay_user_info_share_response", resp.Sign); err != nil {
		return nil, err
	}
	if resp.Error.Code != "" {
		return nil, providerError(resp.Error.Code, firstNonEmpty(resp.Error.SubMsg, resp.Error.Msg, resp.Error.SubCode), raw)
	}
	data := resp.Response
	if !isSuccessCode(data.Code) {
		return nil, providerError(data.Code, firstNonEmpty(data.SubMsg, data.Msg, data.SubCode), raw)
	}

	id := firstNonEmpty(data.OpenID, data.UserID)
	if id == "" {
		return nil, oauth2.ErrInvalidToken
	}
	return &oauth2.User{
		Provider:  providerName,
		ID:        id,
		OpenID:    firstNonEmpty(data.OpenID, data.UserID),
		UnionID:   data.UserID,
		Nickname:  firstNonEmpty(data.NickName, data.UserName),
		AvatarURL: data.Avatar,
		Gender:    data.Gender,
		Province:  data.Province,
		City:      data.City,
		Raw:       raw,
	}, nil
}

func (p *Provider) requestToken(ctx context.Context, values url.Values) (*oauth2.Token, error) {
	raw, err := p.call(ctx, "alipay.system.oauth.token", values)
	if err != nil {
		return nil, err
	}

	var resp tokenEnvelope
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("alipay: decode token: %w", err)
	}
	if err := p.verifyResponse(raw, "alipay_system_oauth_token_response", resp.Sign); err != nil {
		return nil, err
	}
	if resp.Error.Code != "" {
		return nil, providerError(resp.Error.Code, firstNonEmpty(resp.Error.SubMsg, resp.Error.Msg, resp.Error.SubCode), raw)
	}
	data := resp.Response
	if !isSuccessCode(data.Code) {
		return nil, providerError(data.Code, firstNonEmpty(data.SubMsg, data.Msg, data.SubCode), raw)
	}
	if data.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}

	expiresIn := parseExpiresIn(data.ExpiresIn)
	return oauth2.NewToken(oauth2.TokenInfo{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		ExpiresIn:    expiresIn,
		OpenID:       firstNonEmpty(data.OpenID, data.UserID, data.AlipayUserID),
		UnionID:      firstNonEmpty(data.UserID, data.AlipayUserID),
		Raw:          raw,
	}), nil
}

func (p *Provider) call(ctx context.Context, method string, values url.Values) ([]byte, error) {
	params := p.commonParams(method)
	for key, vals := range values {
		for _, val := range vals {
			params.Add(key, val)
		}
	}

	sign, err := signValues(params, p.options.PrivateKey, p.options.SignType)
	if err != nil {
		return nil, err
	}
	params.Set("sign", sign)
	return oauth2.DoPostForm(ctx, p.options.HTTPClient, p.options.GatewayURL, params, providerName)
}

func (p *Provider) commonParams(method string) url.Values {
	values := url.Values{}
	values.Set("app_id", p.options.ClientID)
	values.Set("method", method)
	values.Set("format", p.options.Format)
	values.Set("charset", p.options.Charset)
	values.Set("sign_type", p.options.SignType)
	values.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	values.Set("version", p.options.Version)
	return values
}

func (p *Provider) verifyResponse(raw []byte, responseKey, signature string) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("alipay: decode signed response: %w", err)
	}
	content := envelope[responseKey]
	if len(content) == 0 {
		content = envelope["error_response"]
	}
	if len(content) == 0 {
		return fmt.Errorf("alipay: signed response payload is missing")
	}
	if err := verifyContent(content, signature, p.options.AlipayPublicKey, p.options.SignType); err != nil {
		return fmt.Errorf("alipay: invalid gateway response: %w", err)
	}
	return nil
}

func providerError(code, message string, raw []byte) error {
	return &oauth2.ProviderError{
		Provider: providerName,
		Code:     code,
		Message:  message,
		Raw:      raw,
	}
}

func isSuccessCode(code string) bool {
	return code == "" || code == "10000"
}

func parseExpiresIn(value string) int64 {
	expiresIn, _ := strconv.ParseInt(value, 10, 64)
	return expiresIn
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
	Response tokenResponse `json:"alipay_system_oauth_token_response"`
	Error    errorResponse `json:"error_response"`
	Sign     string        `json:"sign"`
}

type tokenResponse struct {
	Code         string `json:"code"`
	Msg          string `json:"msg"`
	SubCode      string `json:"sub_code"`
	SubMsg       string `json:"sub_msg"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ReExpiresIn  string `json:"re_expires_in"`
	UserID       string `json:"user_id"`
	AlipayUserID string `json:"alipay_user_id"`
	OpenID       string `json:"open_id"`
}

type userEnvelope struct {
	Response userResponse  `json:"alipay_user_info_share_response"`
	Error    errorResponse `json:"error_response"`
	Sign     string        `json:"sign"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

type userResponse struct {
	Code     string `json:"code"`
	Msg      string `json:"msg"`
	SubCode  string `json:"sub_code"`
	SubMsg   string `json:"sub_msg"`
	UserID   string `json:"user_id"`
	OpenID   string `json:"open_id"`
	Avatar   string `json:"avatar"`
	Province string `json:"province"`
	City     string `json:"city"`
	NickName string `json:"nick_name"`
	UserName string `json:"user_name"`
	Gender   string `json:"gender"`
}
