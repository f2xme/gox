package ratelimit

import (
	"errors"
	"time"

	"github.com/f2xme/gox/httpx"
)

// Strategy 定义限流策略
type Strategy int

const (
	// StrategyTokenBucket 使用令牌桶算法（默认）
	// 允许突发流量，最多 burst 个请求
	StrategyTokenBucket Strategy = iota

	// StrategyLeakyBucket 使用漏桶算法
	// 平滑突发流量
	StrategyLeakyBucket

	// StrategyFixedWindow 使用固定窗口算法
	// 简单但可能有边界问题
	StrategyFixedWindow

	// StrategySlidingWindow 使用滑动窗口算法
	// 更精确但占用更多内存
	StrategySlidingWindow
)

// ErrRateLimitExceeded 当超过限流时返回此错误
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// KeyFunc 从上下文中提取限流键
// 常见实现：按 IP、按用户 ID、按 API key 等
type KeyFunc func(ctx httpx.Context) string

// Option 配置限流中间件
type Option func(*Options)

// Options 定义限流中间件配置
type Options struct {
	Strategy Strategy
	Rate     int
	Burst    int
	Window   time.Duration
	KeyFunc  KeyFunc
	Handler  func(ctx httpx.Context)
}

func defaultOptions() *Options {
	return &Options{
		Strategy: StrategyTokenBucket,
		Rate:     100,
		Burst:    100,
		Window:   time.Second,
		KeyFunc:  ByIP,
		Handler:  nil,
	}
}

// WithStrategy 设置限流策略
// 默认为 StrategyTokenBucket
func WithStrategy(s Strategy) Option {
	return func(o *Options) {
		o.Strategy = s
	}
}

// WithRate 设置速率限制（每秒或每窗口的请求数）
// 默认为 100
func WithRate(rate int) Option {
	return func(o *Options) {
		o.Rate = rate
	}
}

// WithBurst 设置令牌桶策略的突发大小
// 默认为 100
func WithBurst(burst int) Option {
	return func(o *Options) {
		o.Burst = burst
	}
}

// WithWindow 设置基于窗口策略的时间窗口
// 默认为 1 秒
func WithWindow(window time.Duration) Option {
	return func(o *Options) {
		o.Window = window
	}
}

// WithKeyFunc 设置从上下文中提取限流键的函数
// 默认为 ByIP
func WithKeyFunc(fn KeyFunc) Option {
	return func(o *Options) {
		o.KeyFunc = fn
	}
}

// WithHandler 设置自定义的限流超出处理函数
// 如果未设置，返回 429 Too Many Requests 和默认消息
func WithHandler(fn func(httpx.Context)) Option {
	return func(o *Options) {
		o.Handler = fn
	}
}

// ByIP 返回按客户端 IP 限流的 KeyFunc
func ByIP(ctx httpx.Context) string {
	return ctx.ClientIP()
}

// ByHeader 返回按指定请求头值限流的 KeyFunc
func ByHeader(header string) KeyFunc {
	return func(ctx httpx.Context) string {
		return ctx.Header(header).String()
	}
}

// ByPath 返回按请求路径限流的 KeyFunc
func ByPath(ctx httpx.Context) string {
	return ctx.Path()
}

// ByIPAndPath 返回按 IP 和路径组合限流的 KeyFunc
func ByIPAndPath(ctx httpx.Context) string {
	return ctx.ClientIP() + ":" + ctx.Path()
}
