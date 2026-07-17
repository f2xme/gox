package aliyun

import (
	"strings"
	"time"
)

const (
	defaultEndpointShanghai = "cloudauth.cn-shanghai.aliyuncs.com"
	defaultEndpointBeijing  = "cloudauth.cn-beijing.aliyuncs.com"
	defaultTimeout          = 5 * time.Second
	paramTypeNormal         = "normal"
)

// Options 阿里云二要素配置。
type Options struct {
	// AccessKeyID 阿里云 AccessKey ID。
	AccessKeyID string
	// AccessKeySecret 阿里云 AccessKey Secret。
	AccessKeySecret string
	// Endpoints Cloudauth 节点列表；空则默认上海、北京。
	Endpoints []string
	// Timeout 单次请求超时。
	Timeout time.Duration
}

// Option 配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Timeout: defaultTimeout,
	}
}

// WithAccessKeyID 设置 AccessKey ID。
func WithAccessKeyID(id string) Option {
	return func(o *Options) { o.AccessKeyID = id }
}

// WithAccessKeySecret 设置 AccessKey Secret。
func WithAccessKeySecret(secret string) Option {
	return func(o *Options) { o.AccessKeySecret = secret }
}

// WithEndpoints 设置 endpoint 列表（主备顺序）。
func WithEndpoints(endpoints ...string) Option {
	return func(o *Options) {
		o.Endpoints = append([]string(nil), endpoints...)
	}
}

// WithTimeout 设置请求超时。
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		if d > 0 {
			o.Timeout = d
		}
	}
}

func (o *Options) normalize() {
	o.AccessKeyID = strings.TrimSpace(o.AccessKeyID)
	o.AccessKeySecret = strings.TrimSpace(o.AccessKeySecret)
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	o.Endpoints = normalizeEndpoints(o.Endpoints)
}

func normalizeEndpoints(endpoints []string) []string {
	out := make([]string, 0, len(endpoints))
	seen := make(map[string]struct{}, len(endpoints))
	for _, ep := range endpoints {
		ep = strings.TrimSpace(ep)
		if ep == "" {
			continue
		}
		if _, ok := seen[ep]; ok {
			continue
		}
		seen[ep] = struct{}{}
		out = append(out, ep)
	}
	if len(out) == 0 {
		return []string{defaultEndpointShanghai, defaultEndpointBeijing}
	}
	return out
}
