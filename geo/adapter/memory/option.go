package memory

import "github.com/f2xme/gox/geo"

// Options 定义内存 IP 地区适配器配置选项。
type Options struct {
	// Locations 初始 IP 到地区信息的映射。
	Locations map[string]*geo.Location
	// LookupError 查询时固定返回的错误，用于测试失败分支。
	LookupError error
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置选项。
func defaultOptions() Options {
	return Options{
		Locations: make(map[string]*geo.Location),
	}
}

// WithLocation 注册一条 IP 地区映射。
//
// 示例：
//
//	New(WithLocation("1.2.3.4", &geo.Location{Country: "中国"}))
func WithLocation(ip string, loc *geo.Location) Option {
	return func(o *Options) {
		if o.Locations == nil {
			o.Locations = make(map[string]*geo.Location)
		}
		if loc == nil {
			o.Locations[ip] = nil
			return
		}
		o.Locations[ip] = loc.Clone()
	}
}

// WithLocations 批量注册 IP 地区映射。
//
// 示例：
//
//	New(WithLocations(map[string]*geo.Location{
//		"1.1.1.1": {Country: "澳大利亚"},
//	}))
func WithLocations(locations map[string]*geo.Location) Option {
	return func(o *Options) {
		if o.Locations == nil {
			o.Locations = make(map[string]*geo.Location)
		}
		for ip, loc := range locations {
			if loc == nil {
				o.Locations[ip] = nil
				continue
			}
			o.Locations[ip] = loc.Clone()
		}
	}
}

// WithLookupError 设置查询时固定返回的错误。
//
// 示例：
//
//	New(WithLookupError(errors.New("lookup failed")))
func WithLookupError(err error) Option {
	return func(o *Options) {
		o.LookupError = err
	}
}
