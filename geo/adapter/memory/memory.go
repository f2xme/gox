package memory

import (
	"context"
	"sync"

	"github.com/f2xme/gox/geo"
)

// Locator 是基于内存的 IP 地区查询实现。
type Locator struct {
	mu        sync.RWMutex
	options   Options
	locations map[string]*geo.Location
}

var _ geo.Locator = (*Locator)(nil)

// Lookup 查询 IP 对应的地区信息。
func (l *Locator) Lookup(ctx context.Context, ip string) (*geo.Location, error) {
	if ctx == nil {
		return nil, geo.NewError(geo.ErrCodeInvalidArgument, "context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, geo.WrapError(geo.ErrCodeInternal, "context error", err, ip)
	}

	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return nil, err
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.options.LookupError != nil {
		return nil, l.options.LookupError
	}

	loc, ok := l.locations[normalized]
	if !ok {
		return nil, geo.NewError(geo.ErrCodeNotFound, "location not found", normalized)
	}
	return loc.Clone(), nil
}

// Set 注册或更新一条 IP 地区映射。
func (l *Locator) Set(ip string, loc *geo.Location) error {
	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return err
	}
	if loc == nil {
		return geo.NewError(geo.ErrCodeInvalidArgument, "location is required", normalized)
	}

	cloned := loc.Clone()
	if cloned.IP == "" {
		cloned.IP = normalized
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.locations[normalized] = cloned
	return nil
}

// Delete 删除一条 IP 地区映射。
func (l *Locator) Delete(ip string) error {
	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.locations, normalized)
	return nil
}

// Reset 清空所有 IP 地区映射。
func (l *Locator) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.locations = make(map[string]*geo.Location)
}

// Count 返回已注册映射数量。
func (l *Locator) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.locations)
}

// SetLookupError 设置查询时固定返回的错误。
func (l *Locator) SetLookupError(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.options.LookupError = err
}
