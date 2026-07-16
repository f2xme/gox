package onepay

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/f2xme/gox/payment"
)

func (s *Service) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, s.config.Path) {
		http.NotFound(w, r)
		return
	}
	token := strings.TrimPrefix(r.URL.Path, s.config.Path)
	if token == "" || strings.Contains(token, "/") {
		s.writeError(w, http.StatusBadRequest, "支付码无效，请重新扫码")
		return
	}
	payload, err := s.decryptToken(token)
	if errors.Is(err, payment.ErrExpired) {
		s.writeError(w, http.StatusGone, "支付码已过期，请获取新码")
		return
	}
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "支付码无效，请重新扫码")
		return
	}
	ua := strings.ToLower(r.UserAgent())
	switch {
	case strings.Contains(ua, "micromessenger"):
		s.handleWechat(w, r, token, payload)
	case strings.Contains(ua, "alipayclient"):
		s.handleAlipay(w, r, payload)
	default:
		s.writeError(w, http.StatusBadRequest, "请使用微信或支付宝扫码")
	}
}

func (s *Service) handleAlipay(w http.ResponseWriter, r *http.Request, payload tokenPayload) {
	checkout, err := s.config.Resolver.ResolveOrCreate(r.Context(), payload.IntentID, payment.ProviderAlipay, "")
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	if !s.validCheckout(checkout, payment.ProviderAlipay) || checkout.WAP == nil || checkout.JSAPI != nil || !allowedAlipayURL(checkout.WAP.URL) {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	http.Redirect(w, r, checkout.WAP.URL, http.StatusSeeOther)
}

func (s *Service) handleWechat(w http.ResponseWriter, r *http.Request, token string, payload tokenPayload) {
	code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
	cookieID := cookieName(token)
	if code == "" {
		expiresAt := time.Unix(payload.ExpiresAt, 0)
		stateValue, statePayload, err := s.createState(token, expiresAt)
		if err != nil {
			s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieID, Value: statePayload.Nonce, Path: r.URL.Path, Expires: expiresAt, MaxAge: int(expiresAt.Sub(s.now()).Seconds()), Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
		redirectURL := s.baseURL.String() + r.URL.Path
		oauthURL, err := s.config.Wechat.OAuthURL(redirectURL, stateValue)
		if err != nil {
			s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
			return
		}
		http.Redirect(w, r, oauthURL, http.StatusSeeOther)
		return
	}
	cookie, err := r.Cookie(cookieID)
	if err != nil || s.verifyState(state, token, cookie.Value) != nil {
		s.deleteStateCookie(w, r.URL.Path, cookieID)
		s.writeError(w, http.StatusBadRequest, "授权状态无效，请重新扫码")
		return
	}
	s.deleteStateCookie(w, r.URL.Path, cookieID)
	openID, err := s.config.Wechat.ExchangeOAuthCode(r.Context(), code)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	checkout, err := s.config.Resolver.ResolveOrCreate(r.Context(), payload.IntentID, payment.ProviderWechat, openID)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	if !s.validCheckout(checkout, payment.ProviderWechat) || checkout.JSAPI == nil || checkout.WAP != nil {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	nonce := s.randomString(18)
	data, err := buildWechatBridgeData(nonce, checkout.JSAPI, s.config.WechatPage)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "支付服务暂时不可用，请稍后重试")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'nonce-"+nonce+"'; base-uri 'none'; frame-ancestors 'none'")
	w.Header().Set("Cache-Control", "no-store")
	_ = wechatBridgeTemplate(s.config.WechatPage).Execute(w, data)
}

func (s *Service) validCheckout(checkout *Checkout, provider payment.Provider) bool {
	return checkout != nil && checkout.Provider == provider && checkout.OrderID != "" && checkout.ExpiresAt.After(s.now())
}

func (s *Service) deleteStateCookie(w http.ResponseWriter, path, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Path: path, MaxAge: -1, Expires: time.Unix(1, 0), Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (s *Service) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = errorTemplate.Execute(w, struct{ Message, RequestID string }{message, s.randomString(9)})
}

func (s *Service) randomString(size int) string {
	raw := make([]byte, size)
	if s.random(raw) != nil {
		return "unavailable"
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func allowedAlipayURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return false
	}
	if u.Port() != "" {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "openapi.alipay.com" || host == "openapi-sandbox.dl.alipaydev.com"
}
