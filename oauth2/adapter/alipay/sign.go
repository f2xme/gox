package alipay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var errUnsupportedSignType = errors.New("alipay: unsupported sign type")

func signValues(values url.Values, privateKey, signType string) (string, error) {
	key, err := parsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	content := signContent(values)
	var hashed []byte
	var hash crypto.Hash
	switch strings.ToUpper(signType) {
	case "RSA":
		sum := sha1.Sum([]byte(content))
		hashed = sum[:]
		hash = crypto.SHA1
	case "RSA2":
		sum := sha256.Sum256([]byte(content))
		hashed = sum[:]
		hash = crypto.SHA256
	default:
		return "", errUnsupportedSignType
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, key, hash, hashed)
	if err != nil {
		return "", fmt.Errorf("alipay: sign request: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func verifyContent(content []byte, signature, publicKey, signType string) error {
	if signature == "" {
		return errors.New("alipay: response signature is empty")
	}
	key, err := parsePublicKey(publicKey)
	if err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("alipay: decode response signature: %w", err)
	}

	var hashed []byte
	var hash crypto.Hash
	switch strings.ToUpper(signType) {
	case "RSA":
		sum := sha1.Sum(content)
		hashed = sum[:]
		hash = crypto.SHA1
	case "RSA2":
		sum := sha256.Sum256(content)
		hashed = sum[:]
		hash = crypto.SHA256
	default:
		return errUnsupportedSignType
	}
	if err := rsa.VerifyPKCS1v15(key, hash, hashed, sig); err != nil {
		return fmt.Errorf("alipay: verify response signature: %w", err)
	}
	return nil
}

func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("alipay: invalid private key pem")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("alipay: parse private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("alipay: private key is not rsa")
	}
	return key, nil
}

func parsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("alipay: invalid public key pem")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("alipay: parse public key: %w", err)
	}
	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("alipay: public key is not rsa")
	}
	return key, nil
}

func signContent(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key, vals := range values {
		if key == "sign" || len(vals) == 0 || vals[0] == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}
	return strings.Join(parts, "&")
}
