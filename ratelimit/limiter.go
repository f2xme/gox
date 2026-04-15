package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// tokenBucket 使用令牌桶算法实现 Limiter 接口
type tokenBucket struct {
	limiter *rate.Limiter
}

// NewTokenBucket 创建一个新的令牌桶限流器
//
// 参数：
//   - rate: 每秒允许的事件数
//   - burst: 最大突发事件数
//
// 示例：
//
//	// 每秒 10 个请求，允许突发 5 个
//	limiter := ratelimit.NewTokenBucket(10, 5)
func NewTokenBucket(r int, burst int) Limiter {
	return &tokenBucket{
		limiter: rate.NewLimiter(rate.Limit(r), burst),
	}
}

func (tb *tokenBucket) Allow() bool {
	return tb.limiter.Allow()
}

func (tb *tokenBucket) Wait(ctx context.Context) error {
	return tb.limiter.Wait(ctx)
}

func (tb *tokenBucket) Reserve() Reservation {
	return &tokenBucketReservation{
		reservation: tb.limiter.Reserve(),
	}
}

// tokenBucketReservation 包装 rate.Reservation 以实现 Reservation 接口
type tokenBucketReservation struct {
	reservation *rate.Reservation
}

func (r *tokenBucketReservation) OK() bool {
	return r.reservation.OK()
}

func (r *tokenBucketReservation) Delay() time.Duration {
	return r.reservation.Delay()
}

func (r *tokenBucketReservation) Cancel() {
	r.reservation.Cancel()
}

// slidingWindow 使用滑动窗口算法实现 Limiter 接口
type slidingWindow struct {
	mu         sync.Mutex
	limit      int
	window     time.Duration
	timestamps []time.Time
}

// NewSlidingWindow 创建一个新的滑动窗口限流器
//
// 参数：
//   - limit: 窗口内允许的最大事件数
//   - window: 时间窗口的持续时间
//
// 示例：
//
//	// 每分钟最多 100 个请求
//	limiter := ratelimit.NewSlidingWindow(100, time.Minute)
func NewSlidingWindow(limit int, window time.Duration) Limiter {
	return &slidingWindow{
		limit:      limit,
		window:     window,
		timestamps: make([]time.Time, 0, limit),
	}
}

func (sw *slidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.cleanup(now)

	if len(sw.timestamps) < sw.limit {
		sw.timestamps = append(sw.timestamps, now)
		return true
	}

	return false
}

func (sw *slidingWindow) Wait(ctx context.Context) error {
	for {
		sw.mu.Lock()
		now := time.Now()
		sw.cleanup(now)

		if len(sw.timestamps) < sw.limit {
			sw.timestamps = append(sw.timestamps, now)
			sw.mu.Unlock()
			return nil
		}

		oldestTime := sw.timestamps[0]
		waitDuration := sw.window - now.Sub(oldestTime)
		sw.mu.Unlock()

		if waitDuration <= 0 {
			waitDuration = time.Microsecond
		}

		timer := time.NewTimer(waitDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (sw *slidingWindow) Reserve() Reservation {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.cleanup(now)

	if len(sw.timestamps) < sw.limit {
		sw.timestamps = append(sw.timestamps, now)
		return &slidingWindowReservation{ok: true}
	}

	oldestTime := sw.timestamps[0]
	delay := max(0, sw.window-now.Sub(oldestTime))

	return &slidingWindowReservation{
		ok:    false,
		delay: delay,
	}
}

// cleanup 从窗口中移除过期的时间戳
func (sw *slidingWindow) cleanup(now time.Time) {
	cutoff := now.Add(-sw.window)

	// 找到第一个未过期的时间戳
	i := 0
	for i < len(sw.timestamps) && !sw.timestamps[i].After(cutoff) {
		i++
	}

	// 移除所有过期的时间戳（直接切片，无需 copy）
	if i > 0 {
		sw.timestamps = sw.timestamps[i:]
	}
}

// slidingWindowReservation 为滑动窗口实现 Reservation 接口
type slidingWindowReservation struct {
	ok    bool
	delay time.Duration
}

func (r *slidingWindowReservation) OK() bool {
	return r.ok
}

func (r *slidingWindowReservation) Delay() time.Duration {
	return r.delay
}

// Cancel 对于滑动窗口不执行任何操作，因为它不支持取消预留
func (r *slidingWindowReservation) Cancel() {
}
