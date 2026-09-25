package x

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const pkceVerifierBytes = 32

type codeVerifierKey struct{}

// NewPKCE 生成一对 PKCE verifier 和 S256 challenge。
//
// verifier 必须由业务侧随 state 一起保存，换票时再交回。不要把它放到授权地址或浏览器日志里。
func NewPKCE() (verifier, challenge string, err error) {
	buf := make([]byte, pkceVerifierBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("x: generate pkce verifier: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	return verifier, s256Challenge(verifier), nil
}

// WithCodeVerifier 把 PKCE verifier 放入 context，供 Provider.Exchange 读取。
//
// Provider.Exchange 的签名只有授权码。X 换票还必须提交本次授权使用的 verifier。
func WithCodeVerifier(ctx context.Context, verifier string) context.Context {
	return context.WithValue(ctx, codeVerifierKey{}, verifier)
}

func codeVerifierFrom(ctx context.Context) string {
	verifier, _ := ctx.Value(codeVerifierKey{}).(string)
	return verifier
}

func s256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
