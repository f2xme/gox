package memory

import (
	"log"

	"github.com/f2xme/gox/geo"
)

// New 创建新的内存 IP 地区适配器。
func New(opts ...Option) (*Locator, error) {
	options := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	locations := make(map[string]*geo.Location, len(options.Locations))
	for ip, loc := range options.Locations {
		normalized, err := geo.NormalizeIP(ip)
		if err != nil {
			return nil, err
		}
		if loc == nil {
			continue
		}
		cloned := loc.Clone()
		if cloned.IP == "" {
			cloned.IP = normalized
		}
		locations[normalized] = cloned
	}

	return &Locator{
		options:   options,
		locations: locations,
	}, nil
}

// MustNew 创建新的内存 IP 地区适配器，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Locator {
	locator, err := New(opts...)
	if err != nil {
		log.Fatalf("memory geo: create locator failed: %v", err)
	}
	return locator
}
