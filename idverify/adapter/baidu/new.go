package baidu

import (
	"fmt"
	"strings"

	"github.com/f2xme/gox/idverify"
)

// New 创建百度二要素核验器。
func New(opts ...Option) (*Verifier, error) {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	o.APIKey = strings.TrimSpace(o.APIKey)
	o.SecretKey = strings.TrimSpace(o.SecretKey)
	if err := o.validate(); err != nil {
		return nil, err
	}
	return &Verifier{options: o}, nil
}

// MustNew 创建百度核验器，失败 panic。
func MustNew(opts ...Option) *Verifier {
	v, err := New(opts...)
	if err != nil {
		panic(fmt.Errorf("baidu idverify: %w", err))
	}
	return v
}

func idverifyNotConfigured() error {
	return idverify.Wrap(idverify.ProviderBaidu, "config", idverify.ErrNotConfigured)
}
