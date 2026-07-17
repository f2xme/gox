package aliyun

import (
	"fmt"

	"github.com/f2xme/gox/idverify"
)

// New 创建阿里云二要素核验器。
func New(opts ...Option) (*Verifier, error) {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	o.normalize()
	if o.AccessKeyID == "" || o.AccessKeySecret == "" {
		return nil, idverify.Wrap(idverify.ProviderAliyun, "config", idverify.ErrNotConfigured)
	}
	return &Verifier{
		options: o,
	}, nil
}

// MustNew 创建阿里云核验器，失败 panic。
func MustNew(opts ...Option) *Verifier {
	v, err := New(opts...)
	if err != nil {
		panic(fmt.Errorf("aliyun idverify: %w", err))
	}
	return v
}
