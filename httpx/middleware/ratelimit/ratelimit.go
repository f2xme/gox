package ratelimit

import (
	"sync"
	"time"

	"github.com/f2xme/gox/httpx"
)

// New 创建限流中间件
// 默认策略为令牌桶，每秒 100 个请求
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var limiter Limiter
	switch o.Strategy {
	case StrategyTokenBucket:
		limiter = newTokenBucket(o.Rate, o.Burst)
	case StrategyLeakyBucket:
		limiter = newLeakyBucket(o.Rate)
	case StrategyFixedWindow:
		limiter = newFixedWindow(o.Rate, o.Window)
	case StrategySlidingWindow:
		limiter = newSlidingWindow(o.Rate, o.Window)
	default:
		limiter = newTokenBucket(o.Rate, o.Burst)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			// Extract key for rate limiting
			key := o.KeyFunc(ctx)

			// Check rate limit
			allowed := limiter.Allow(key)
			if !allowed {
				if o.Handler != nil {
					o.Handler(ctx)
					return ErrRateLimitExceeded
				}

				// Default rate limit response
				ctx.Status(429)
				ctx.JSON(429, map[string]any{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded, please try again later",
				})
				return ErrRateLimitExceeded
			}

			return next(ctx)
		}
	}
}

// Limiter 定义限流接口
type Limiter interface {
	Allow(key string) bool
}

// tokenBucket 实现令牌桶算法
type tokenBucket struct {
	rate     int
	burst    int
	buckets  map[string]*bucket
	mu       sync.RWMutex
	cleanupT *time.Ticker
	stopCh   chan struct{}
}

type bucket struct {
	tokens    float64
	lastRefill time.Time
	mu        sync.Mutex
}

func newTokenBucket(rate, burst int) *tokenBucket {
	tb := &tokenBucket{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*bucket),
		stopCh:  make(chan struct{}),
	}

	// Cleanup expired buckets every minute
	tb.cleanupT = time.NewTicker(time.Minute)
	go tb.cleanup()

	return tb
}

func (tb *tokenBucket) Allow(key string) bool {
	tb.mu.RLock()
	b, exists := tb.buckets[key]
	tb.mu.RUnlock()

	if !exists {
		tb.mu.Lock()
		b = &bucket{
			tokens:    float64(tb.burst),
			lastRefill: time.Now(),
		}
		tb.buckets[key] = b
		tb.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	// Refill tokens
	b.tokens += elapsed * float64(tb.rate)
	if b.tokens > float64(tb.burst) {
		b.tokens = float64(tb.burst)
	}
	b.lastRefill = now

	// Check if we have tokens
	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func (tb *tokenBucket) cleanup() {
	for {
		select {
		case <-tb.cleanupT.C:
			tb.mu.Lock()
			now := time.Now()
			for key, b := range tb.buckets {
				b.mu.Lock()
				if now.Sub(b.lastRefill) > 5*time.Minute {
					delete(tb.buckets, key)
				}
				b.mu.Unlock()
			}
			tb.mu.Unlock()
		case <-tb.stopCh:
			tb.cleanupT.Stop()
			return
		}
	}
}

// leakyBucket 实现漏桶算法
type leakyBucket struct {
	rate    int
	buckets map[string]*leakyBucketState
	mu      sync.RWMutex
}

type leakyBucketState struct {
	lastLeak time.Time
	count    int
	mu       sync.Mutex
}

func newLeakyBucket(rate int) *leakyBucket {
	return &leakyBucket{
		rate:    rate,
		buckets: make(map[string]*leakyBucketState),
	}
}

func (lb *leakyBucket) Allow(key string) bool {
	lb.mu.RLock()
	state, exists := lb.buckets[key]
	lb.mu.RUnlock()

	if !exists {
		lb.mu.Lock()
		state = &leakyBucketState{
			lastLeak: time.Now(),
			count:    0,
		}
		lb.buckets[key] = state
		lb.mu.Unlock()
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(state.lastLeak).Seconds()

	// Leak tokens
	leaked := int(elapsed * float64(lb.rate))
	state.count -= leaked
	if state.count < 0 {
		state.count = 0
	}
	state.lastLeak = now

	// Check capacity
	if state.count < lb.rate {
		state.count++
		return true
	}

	return false
}

// fixedWindow 实现固定窗口算法
type fixedWindow struct {
	rate    int
	window  time.Duration
	windows map[string]*windowState
	mu      sync.RWMutex
}

type windowState struct {
	count      int
	windowStart time.Time
	mu         sync.Mutex
}

func newFixedWindow(rate int, window time.Duration) *fixedWindow {
	return &fixedWindow{
		rate:    rate,
		window:  window,
		windows: make(map[string]*windowState),
	}
}

func (fw *fixedWindow) Allow(key string) bool {
	fw.mu.RLock()
	state, exists := fw.windows[key]
	fw.mu.RUnlock()

	if !exists {
		fw.mu.Lock()
		state = &windowState{
			count:      0,
			windowStart: time.Now(),
		}
		fw.windows[key] = state
		fw.mu.Unlock()
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Check if window expired
	if now.Sub(state.windowStart) >= fw.window {
		state.count = 0
		state.windowStart = now
	}

	// Check rate limit
	if state.count < fw.rate {
		state.count++
		return true
	}

	return false
}

// slidingWindow 实现滑动窗口算法
type slidingWindow struct {
	rate    int
	window  time.Duration
	windows map[string]*slidingWindowState
	mu      sync.RWMutex
}

type slidingWindowState struct {
	requests []time.Time
	mu       sync.Mutex
}

func newSlidingWindow(rate int, window time.Duration) *slidingWindow {
	return &slidingWindow{
		rate:    rate,
		window:  window,
		windows: make(map[string]*slidingWindowState),
	}
}

func (sw *slidingWindow) Allow(key string) bool {
	sw.mu.RLock()
	state, exists := sw.windows[key]
	sw.mu.RUnlock()

	if !exists {
		sw.mu.Lock()
		state = &slidingWindowState{
			requests: make([]time.Time, 0),
		}
		sw.windows[key] = state
		sw.mu.Unlock()
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// Remove expired requests
	i := 0
	for i < len(state.requests) && !state.requests[i].After(cutoff) {
		i++
	}
	if i > 0 {
		n := copy(state.requests, state.requests[i:])
		state.requests = state.requests[:n]
	}

	// Check rate limit
	if len(state.requests) < sw.rate {
		state.requests = append(state.requests, now)
		return true
	}

	return false
}
