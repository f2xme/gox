package baidu

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/f2xme/gox/geo"
)

// New 创建新的百度 IP 查询适配器。
func New(opts ...Option) (*Locator, error) {
	options := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	// 与 http/amap 适配器一致：<=0 回落默认 5s，避免无限挂起
	if options.Timeout <= 0 {
		options.Timeout = 5 * time.Second
	}

	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: options.Timeout}
	} else if client.Timeout == 0 {
		cloned := *client
		cloned.Timeout = options.Timeout
		client = &cloned
	}

	return &Locator{
		endpoint: strings.TrimSpace(options.Endpoint),
		client:   client,
	}, nil
}

// MustNew 创建新的百度 IP 查询适配器，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Locator {
	locator, err := New(opts...)
	if err != nil {
		log.Fatalf("baidu geo: create locator failed: %v", err)
	}
	return locator
}

func validateOptions(o Options) error {
	endpoint := strings.TrimSpace(o.Endpoint)
	if endpoint == "" {
		return geo.NewError(geo.ErrCodeInvalidArgument, "endpoint is required")
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return geo.NewError(geo.ErrCodeInvalidArgument, "endpoint must start with http:// or https://")
	}
	if o.Timeout < 0 {
		return geo.NewError(geo.ErrCodeInvalidArgument, "timeout must not be negative")
	}
	return nil
}
