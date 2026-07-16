package onepay

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/f2xme/gox/payment"
)

// Code 是可展示的一码付 URL 与 PNG。
type Code struct {
	// URL 是二维码编码的中立支付地址。
	URL string
	// PNG 是二维码图片内容。
	PNG []byte
	// ExpiresAt 是二维码过期时间。
	ExpiresAt time.Time
}

// Service 生成中立二维码并提供扫码 HTTP handler。
type Service struct {
	config  Config
	baseURL *url.URL
	qrSize  int
	encode  qrEncoder
	now     func() time.Time
	random  func([]byte) error
	handler http.Handler
}

// New 创建一码付服务。
func New(config Config, opts ...Option) (*Service, error) {
	c, baseURL, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}
	o := options{qrSize: 256}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.qrSize < 128 || o.qrSize > 1024 {
		return nil, fmt.Errorf("%w: QR size must be between 128 and 1024", payment.ErrInvalidConfig)
	}
	s := &Service{config: c, baseURL: baseURL, qrSize: o.qrSize, encode: encodePNG, now: time.Now, random: func(b []byte) error { _, err := rand.Read(b); return err }}
	s.handler = http.HandlerFunc(s.serveHTTP)
	return s, nil
}

// CreateCode 生成中立 URL 与内存 PNG。
func (s *Service) CreateCode(ctx context.Context, intentID string, opts ...CodeOption) (*Code, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if intentID == "" {
		return nil, fmt.Errorf("%w: intent ID cannot be empty", payment.ErrInvalidRequest)
	}
	now := s.now()
	o := codeOptions{expiresAt: now.Add(s.config.TokenTTL)}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	issuedAtUnix, expiresAtUnix := now.Unix(), o.expiresAt.Unix()
	if expiresAtUnix <= issuedAtUnix || o.expiresAt.After(now.Add(maxTTL)) {
		return nil, fmt.Errorf("%w: code expiration must be within 24 hours", payment.ErrInvalidRequest)
	}
	token, err := s.encryptToken(tokenPayload{Version: 1, IntentID: intentID, IssuedAt: issuedAtUnix, ExpiresAt: expiresAtUnix})
	if err != nil {
		return nil, fmt.Errorf("create onepay token: %w", err)
	}
	rawURL := s.baseURL.String() + s.config.Path + token
	png, err := s.encode(rawURL, s.qrSize)
	if err != nil {
		return nil, fmt.Errorf("encode onepay QR: %w", err)
	}
	return &Code{URL: rawURL, PNG: png, ExpiresAt: o.expiresAt}, nil
}

// Handler 返回只处理 Config.Path 前缀的 HTTP handler。
func (s *Service) Handler() http.Handler { return s.handler }
