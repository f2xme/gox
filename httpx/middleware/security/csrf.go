package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/f2xme/gox/httpx"
)

const (
	// CSRFTokenContextKey 是存储 CSRF token 的上下文键
	CSRFTokenContextKey = "security.csrf_token"
)

// GetCSRFToken 从请求上下文中返回当前的 CSRF token
func GetCSRFToken(ctx httpx.Context) string {
	if token, ok := ctx.Get(CSRFTokenContextKey); ok {
		if typed, ok := token.(string); ok {
			return typed
		}
	}
	return ""
}

func ensureCSRFCookie(ctx httpx.Context, cfg *CSRFConfig) (string, error) {
	if cookie, err := ctx.Cookie(cfg.CookieName); err == nil && cookie.Value != "" {
		ctx.Set(CSRFTokenContextKey, cookie.Value)
		return cookie.Value, nil
	}

	token, err := generateCSRFToken(cfg.TokenLength)
	if err != nil {
		return "", err
	}
	ctx.SetCookie(&http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     cfg.CookiePath,
		MaxAge:   cfg.CookieMaxAge,
		HttpOnly: false, // Must be false so JavaScript can read the token for CSRF protection
		Secure:   cfg.CookieSecure,
		SameSite: parseSameSite(cfg.CookieSameSite),
	})
	ctx.Set(CSRFTokenContextKey, token)
	return token, nil
}

func validateCSRFToken(ctx httpx.Context, cfg *CSRFConfig) bool {
	cookie, err := ctx.Cookie(cfg.CookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	requestToken := lookupCSRFToken(ctx, cfg.TokenLookup)
	if requestToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(requestToken)) == 1
}

func generateCSRFToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func lookupCSRFToken(ctx httpx.Context, tokenLookup string) string {
	parts := strings.SplitN(tokenLookup, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	switch parts[0] {
	case "header":
		return ctx.Header(parts[1]).String()
	case "query":
		return ctx.Query(parts[1]).String()
	default:
		return ""
	}
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
