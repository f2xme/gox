package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/f2xme/gox/httpx"
	"golang.org/x/time/rate"
)

const limiterIdleTTL = 5 * time.Minute

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
			key := o.KeyFunc(ctx)
			if !limiter.Allow(key) {
				if o.Handler != nil {
					o.Handler(ctx)
					return ErrRateLimitExceeded
				}

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

// keyedStore 保存每个限流键的独立状态，并在请求到来时清理长期未使用的键。
// 惰性清理避免为每个中间件启动常驻 goroutine。
type keyedStore[T any] struct {
	mu          sync.RWMutex
	values      map[string]*keyedValue[T]
	newValue    func() T
	idleTTL     time.Duration
	nextCleanup atomic.Int64
}

type keyedValue[T any] struct {
	value    T
	lastUsed atomic.Int64
}

func newKeyedStore[T any](idleTTL time.Duration, newValue func() T) *keyedStore[T] {
	store := &keyedStore[T]{
		values:   make(map[string]*keyedValue[T]),
		newValue: newValue,
		idleTTL:  idleTTL,
	}
	store.nextCleanup.Store(time.Now().Add(idleTTL).UnixNano())
	return store
}

func (s *keyedStore[T]) get(key string) T {
	now := time.Now()

	s.mu.RLock()
	storedValue, ok := s.values[key]
	if ok {
		storedValue.lastUsed.Store(now.UnixNano())
	}
	s.mu.RUnlock()

	if !ok {
		s.mu.Lock()
		storedValue, ok = s.values[key]
		if !ok {
			storedValue = &keyedValue[T]{value: s.newValue()}
			s.values[key] = storedValue
		}
		storedValue.lastUsed.Store(now.UnixNano())
		s.mu.Unlock()
	}

	s.cleanup(now)
	return storedValue.value
}

func (s *keyedStore[T]) cleanup(now time.Time) {
	nextCleanup := s.nextCleanup.Load()
	if now.UnixNano() < nextCleanup ||
		!s.nextCleanup.CompareAndSwap(nextCleanup, now.Add(s.idleTTL).UnixNano()) {
		return
	}

	idleBefore := now.Add(-s.idleTTL).UnixNano()
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range s.values {
		if value.lastUsed.Load() <= idleBefore {
			delete(s.values, key)
		}
	}
}

// tokenBucket 使用 x/time/rate 实现令牌桶算法。
type tokenBucket struct {
	buckets *keyedStore[*rate.Limiter]
}

func newTokenBucket(requestsPerSecond, burst int) *tokenBucket {
	idleTTL := limiterIdleTTL
	if requestsPerSecond > 0 && burst > 0 {
		refillDuration := time.Duration(float64(burst) / float64(requestsPerSecond) * float64(time.Second))
		idleTTL = max(idleTTL, refillDuration)
	}

	return &tokenBucket{
		buckets: newKeyedStore(idleTTL, func() *rate.Limiter {
			return rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
		}),
	}
}

func (tb *tokenBucket) Allow(key string) bool {
	return tb.buckets.get(key).Allow()
}

// leakyBucket 实现漏桶算法
type leakyBucket struct {
	rate    int
	buckets *keyedStore[*leakyBucketState]
}

type leakyBucketState struct {
	lastLeak time.Time
	level    float64
	mu       sync.Mutex
}

func newLeakyBucket(rate int) *leakyBucket {
	return &leakyBucket{
		rate: rate,
		buckets: newKeyedStore(limiterIdleTTL, func() *leakyBucketState {
			return &leakyBucketState{lastLeak: time.Now()}
		}),
	}
}

func (lb *leakyBucket) Allow(key string) bool {
	state := lb.buckets.get(key)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(state.lastLeak).Seconds()
	state.level = max(0, state.level-elapsed*float64(lb.rate))
	state.lastLeak = now

	if state.level+1 <= float64(lb.rate) {
		state.level++
		return true
	}

	return false
}

// fixedWindow 实现固定窗口算法
type fixedWindow struct {
	rate    int
	window  time.Duration
	windows *keyedStore[*windowState]
}

type windowState struct {
	count       int
	windowStart time.Time
	mu          sync.Mutex
}

func newFixedWindow(rate int, window time.Duration) *fixedWindow {
	return &fixedWindow{
		rate:   rate,
		window: window,
		windows: newKeyedStore(max(limiterIdleTTL, window), func() *windowState {
			return &windowState{windowStart: time.Now()}
		}),
	}
}

func (fw *fixedWindow) Allow(key string) bool {
	state := fw.windows.get(key)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	if now.Sub(state.windowStart) >= fw.window {
		state.count = 0
		state.windowStart = now
	}

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
	windows *keyedStore[*slidingWindowState]
}

type slidingWindowState struct {
	requests []time.Time
	mu       sync.Mutex
}

func newSlidingWindow(rate int, window time.Duration) *slidingWindow {
	return &slidingWindow{
		rate:   rate,
		window: window,
		windows: newKeyedStore(max(limiterIdleTTL, window), func() *slidingWindowState {
			return &slidingWindowState{requests: make([]time.Time, 0, rate)}
		}),
	}
}

func (sw *slidingWindow) Allow(key string) bool {
	state := sw.windows.get(key)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	i := 0
	for i < len(state.requests) && !state.requests[i].After(cutoff) {
		i++
	}
	if i > 0 {
		n := copy(state.requests, state.requests[i:])
		state.requests = state.requests[:n]
	}

	if len(state.requests) < sw.rate {
		state.requests = append(state.requests, now)
		return true
	}

	return false
}
