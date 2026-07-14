package http

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/f2xme/gox/geo"
)

// New 创建新的 HTTP IP 地区适配器。
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
	// WithTimeout(0) 或未覆盖时回落到默认超时（defaultOptions 已是 5s；
	// 若调用方显式传入 0，仍视为“使用默认”而非无限等待）。
	if options.Timeout <= 0 {
		options.Timeout = 5 * time.Second
	}

	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: options.Timeout}
	} else if client.Timeout == 0 {
		// 复制客户端配置，避免修改调用方传入的 Client
		cloned := *client
		cloned.Timeout = options.Timeout
		client = &cloned
	}

	parser := options.Parser
	if parser == nil {
		parser = defaultJSONParser
	}

	headers := make(map[string]string, len(options.Headers))
	for k, v := range options.Headers {
		headers[k] = v
	}

	return &Locator{
		endpoint: strings.TrimSpace(options.Endpoint),
		client:   client,
		headers:  headers,
		parser:   parser,
	}, nil
}

// MustNew 创建新的 HTTP IP 地区适配器，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Locator {
	locator, err := New(opts...)
	if err != nil {
		log.Fatalf("http geo: create locator failed: %v", err)
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
