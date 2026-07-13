package onepay

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/f2xme/gox/payment"
)

type tokenPayload struct {
	Version   int    `json:"v"`
	IntentID  string `json:"i"`
	IssuedAt  int64  `json:"n"`
	ExpiresAt int64  `json:"e"`
}

func (s *Service) encryptToken(payload tokenPayload) (string, error) {
	gcm, err := newGCM(s.config.TokenKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if err := s.random(nonce); err != nil {
		return "", err
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plain, []byte("gox-onepay-v1"))
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (s *Service) decryptToken(token string) (tokenPayload, error) {
	var payload tokenPayload
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return payload, payment.ErrInvalidRequest
	}
	gcm, err := newGCM(s.config.TokenKey)
	if err != nil || len(raw) < gcm.NonceSize() {
		return payload, payment.ErrInvalidRequest
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], []byte("gox-onepay-v1"))
	if err != nil || json.Unmarshal(plain, &payload) != nil || payload.Version != 1 || payload.IntentID == "" {
		return tokenPayload{}, payment.ErrInvalidRequest
	}
	now := s.now()
	if time.Unix(payload.IssuedAt, 0).After(now.Add(time.Minute)) {
		return tokenPayload{}, payment.ErrInvalidRequest
	}
	issuedAt, expiresAt := time.Unix(payload.IssuedAt, 0), time.Unix(payload.ExpiresAt, 0)
	if !expiresAt.After(issuedAt) || expiresAt.After(issuedAt.Add(maxTTL)) {
		return tokenPayload{}, payment.ErrInvalidRequest
	}
	if !expiresAt.After(now) {
		return tokenPayload{}, payment.ErrExpired
	}
	return payload, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES: %w", err)
	}
	return cipher.NewGCM(block)
}
