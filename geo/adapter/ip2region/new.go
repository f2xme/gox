package ip2region

import (
	"fmt"
	"log"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/service"

	"github.com/f2xme/gox/geo"
)

// New 创建新的 ip2region 适配器。
//
// 至少需要通过 WithV4DBPath 或 WithV6DBPath 配置一个 xdb 文件。
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

	policy, err := toServiceCachePolicy(options.CachePolicy)
	if err != nil {
		return nil, geo.NewError(geo.ErrCodeInvalidArgument, err.Error())
	}

	var v4Config *service.Config
	if path := strings.TrimSpace(options.V4DBPath); path != "" {
		v4Config, err = service.NewV4Config(policy, path, options.PoolSize)
		if err != nil {
			return nil, geo.WrapError(geo.ErrCodeInvalidArgument, "create v4 config failed", err)
		}
	}

	var v6Config *service.Config
	if path := strings.TrimSpace(options.V6DBPath); path != "" {
		v6Config, err = service.NewV6Config(policy, path, options.PoolSize)
		if err != nil {
			return nil, geo.WrapError(geo.ErrCodeInvalidArgument, "create v6 config failed", err)
		}
	}

	svc, err := service.NewIp2Region(v4Config, v6Config)
	if err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "create ip2region service failed", err)
	}

	return &Locator{
		service: svc,
		options: options,
	}, nil
}

// MustNew 创建新的 ip2region 适配器，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Locator {
	locator, err := New(opts...)
	if err != nil {
		log.Fatalf("ip2region: create locator failed: %v", err)
	}
	return locator
}

func validateOptions(o Options) error {
	if strings.TrimSpace(o.V4DBPath) == "" && strings.TrimSpace(o.V6DBPath) == "" {
		return geo.NewError(geo.ErrCodeInvalidArgument, "at least one of v4 or v6 db path is required")
	}
	if o.PoolSize <= 0 {
		return geo.NewError(geo.ErrCodeInvalidArgument, "pool size must be positive")
	}
	switch o.CachePolicy {
	case CachePolicyNone, CachePolicyVIndex, CachePolicyBuffer, "":
	default:
		return geo.NewError(geo.ErrCodeInvalidArgument, "unsupported cache policy: "+string(o.CachePolicy))
	}
	return nil
}

func toServiceCachePolicy(policy CachePolicy) (int, error) {
	switch policy {
	case CachePolicyNone:
		return service.NoCache, nil
	case CachePolicyVIndex, "":
		return service.VIndexCache, nil
	case CachePolicyBuffer:
		return service.BufferCache, nil
	default:
		return 0, fmt.Errorf("unsupported cache policy: %s", policy)
	}
}
