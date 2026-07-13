package onepay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	stateVersion       = byte(1)
	stateTokenHashSize = 12
	stateNonceSize     = 12
	statePayloadSize   = 1 + stateTokenHashSize + stateNonceSize + 8
	stateTagSize       = 16
	stateRawSize       = statePayloadSize + stateTagSize
)

var stateEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type statePayload struct{ Nonce string }

func (s *Service) createState(token string, expiresAt time.Time) (string, statePayload, error) {
	raw := make([]byte, stateRawSize)
	raw[0] = stateVersion
	copy(raw[1:1+stateTokenHashSize], tokenHash(token))
	nonce := raw[1+stateTokenHashSize : 1+stateTokenHashSize+stateNonceSize]
	if err := s.random(nonce); err != nil {
		return "", statePayload{}, err
	}
	binary.BigEndian.PutUint64(raw[statePayloadSize-8:statePayloadSize], uint64(expiresAt.Unix()))
	tag := s.stateMAC(raw[:statePayloadSize])
	copy(raw[statePayloadSize:], tag[:stateTagSize])
	return stateEncoding.EncodeToString(raw), statePayload{Nonce: stateEncoding.EncodeToString(nonce)}, nil
}

func (s *Service) verifyState(value, token, cookieNonce string) error {
	raw, err := stateEncoding.DecodeString(value)
	if err != nil || len(raw) != stateRawSize || raw[0] != stateVersion {
		return payment.ErrInvalidOAuthState
	}
	tag := s.stateMAC(raw[:statePayloadSize])
	if !hmac.Equal(raw[statePayloadSize:], tag[:stateTagSize]) {
		return payment.ErrInvalidOAuthState
	}
	if !hmac.Equal(raw[1:1+stateTokenHashSize], tokenHash(token)) {
		return payment.ErrInvalidOAuthState
	}
	cookieRaw, err := stateEncoding.DecodeString(cookieNonce)
	stateNonce := raw[1+stateTokenHashSize : 1+stateTokenHashSize+stateNonceSize]
	if err != nil || len(cookieRaw) != stateNonceSize || !hmac.Equal(stateNonce, cookieRaw) {
		return payment.ErrInvalidOAuthState
	}
	expiresAt := int64(binary.BigEndian.Uint64(raw[statePayloadSize-8 : statePayloadSize]))
	if !time.Unix(expiresAt, 0).After(s.now()) {
		return payment.ErrInvalidOAuthState
	}
	return nil
}

func (s *Service) stateMAC(value []byte) []byte {
	derived := sha256.Sum256(append([]byte("gox-onepay-oauth-state:"), s.config.TokenKey...))
	mac := hmac.New(sha256.New, derived[:])
	_, _ = mac.Write(value)
	return mac.Sum(nil)
}

func tokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:stateTokenHashSize]
}

func hashToken(token string) string { return stateEncoding.EncodeToString(tokenHash(token)) }

func cookieName(token string) string { return "onepay_state_" + hashToken(token) }
