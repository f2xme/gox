package ip2region

// CachePolicy 定义 xdb 缓存策略。
type CachePolicy string

const (
	// CachePolicyNone 不缓存，每次查询走文件 IO。
	CachePolicyNone CachePolicy = "none"
	// CachePolicyVIndex 缓存 VectorIndex，兼顾性能与内存（推荐默认）。
	CachePolicyVIndex CachePolicy = "vindex"
	// CachePolicyBuffer 将整个 xdb 载入内存。
	CachePolicyBuffer CachePolicy = "buffer"
)

// Options 定义 ip2region 适配器配置选项。
type Options struct {
	// V4DBPath IPv4 xdb 文件路径。
	V4DBPath string
	// V6DBPath IPv6 xdb 文件路径。
	V6DBPath string
	// CachePolicy 缓存策略，默认 CachePolicyVIndex。
	CachePolicy CachePolicy
	// PoolSize 查询 searcher 池大小，默认 20。
	// BufferCache 策略下该值会被底层忽略。
	PoolSize int
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		CachePolicy: CachePolicyVIndex,
		PoolSize:    20,
	}
}

// WithV4DBPath 设置 IPv4 xdb 文件路径。
//
// 示例：
//
//	New(WithV4DBPath("/data/ip2region_v4.xdb"))
func WithV4DBPath(path string) Option {
	return func(o *Options) {
		o.V4DBPath = path
	}
}

// WithV6DBPath 设置 IPv6 xdb 文件路径。
//
// 示例：
//
//	New(WithV6DBPath("/data/ip2region_v6.xdb"))
func WithV6DBPath(path string) Option {
	return func(o *Options) {
		o.V6DBPath = path
	}
}

// WithCachePolicy 设置缓存策略。
//
// 示例：
//
//	New(WithV4DBPath("v4.xdb"), WithCachePolicy(CachePolicyBuffer))
func WithCachePolicy(policy CachePolicy) Option {
	return func(o *Options) {
		o.CachePolicy = policy
	}
}

// WithPoolSize 设置 searcher 池大小。
//
// 示例：
//
//	New(WithV4DBPath("v4.xdb"), WithPoolSize(32))
func WithPoolSize(size int) Option {
	return func(o *Options) {
		o.PoolSize = size
	}
}
