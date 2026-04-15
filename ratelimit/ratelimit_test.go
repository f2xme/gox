package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestTokenBucketInitialTokens 测试令牌桶创建后初始令牌数正确
func TestTokenBucketInitialTokens(t *testing.T) {
	tests := []struct {
		name  string
		rate  int
		burst int
		want  int // 期望初始可用的令牌数
	}{
		{"burst_5", 10, 5, 5},
		{"burst_1", 10, 1, 1},
		{"burst_10", 100, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewTokenBucket(tt.rate, tt.burst)

			// 初始应该有 burst 个令牌可用
			allowed := 0
			for i := 0; i < tt.want; i++ {
				if limiter.Allow() {
					allowed++
				}
			}

			if allowed != tt.want {
				t.Errorf("initial tokens = %d, want %d", allowed, tt.want)
			}

			// 下一个请求应该被限流
			if limiter.Allow() {
				t.Error("expected rate limit after burst exhausted")
			}
		})
	}
}

// TestTokenBucketAllow 测试 Allow() 在限流时返回 false，未限流时返回 true
func TestTokenBucketAllow(t *testing.T) {
	tests := []struct {
		name      string
		rate      int
		burst     int
		requests  int
		sleepTime time.Duration
		wantAllow int
	}{
		{
			name:      "within_burst",
			rate:      10,
			burst:     5,
			requests:  5,
			sleepTime: 0,
			wantAllow: 5,
		},
		{
			name:      "exceed_burst",
			rate:      10,
			burst:     5,
			requests:  10,
			sleepTime: 0,
			wantAllow: 5,
		},
		{
			name:      "refill_after_sleep",
			rate:      10,
			burst:     2,
			requests:  3,
			sleepTime: 150 * time.Millisecond, // 允许补充 1-2 个令牌
			wantAllow: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewTokenBucket(tt.rate, tt.burst)

			// 先消耗 burst
			for i := 0; i < tt.burst; i++ {
				limiter.Allow()
			}

			if tt.sleepTime > 0 {
				time.Sleep(tt.sleepTime)
			}

			allowed := 0
			for i := 0; i < tt.requests-tt.burst; i++ {
				if limiter.Allow() {
					allowed++
				}
			}

			if tt.sleepTime > 0 && allowed == 0 {
				t.Error("expected tokens to refill after sleep")
			}
		})
	}
}

// TestTokenBucketWait 测试 Wait() 在限流时阻塞，context 取消时返回错误
func TestTokenBucketWait(t *testing.T) {
	tests := []struct {
		name        string
		rate        int
		burst       int
		ctxTimeout  time.Duration
		expectError bool
	}{
		{
			name:        "wait_success",
			rate:        10,
			burst:       1,
			ctxTimeout:  500 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "context_cancelled",
			rate:        1,
			burst:       1,
			ctxTimeout:  10 * time.Millisecond,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewTokenBucket(tt.rate, tt.burst)

			// 消耗初始令牌
			limiter.Allow()

			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			err := limiter.Wait(ctx)

			if tt.expectError && err == nil {
				t.Error("expected error when context cancelled")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestTokenBucketRefill 测试令牌桶按速率正确补充令牌
func TestTokenBucketRefill(t *testing.T) {
	rate := 10 // 每秒 10 个令牌
	burst := 2
	limiter := NewTokenBucket(rate, burst)

	// 消耗所有初始令牌
	for i := 0; i < burst; i++ {
		if !limiter.Allow() {
			t.Fatalf("failed to consume initial token %d", i)
		}
	}

	// 此时应该被限流
	if limiter.Allow() {
		t.Error("expected rate limit after burst exhausted")
	}

	// 等待足够时间补充 1 个令牌 (100ms = 0.1s, rate=10 => 1 token)
	time.Sleep(150 * time.Millisecond)

	// 现在应该有 1 个令牌可用
	if !limiter.Allow() {
		t.Error("expected token to be refilled")
	}

	// 再次应该被限流
	if limiter.Allow() {
		t.Error("expected rate limit after consuming refilled token")
	}
}

// TestTokenBucketConcurrent 测试并发调用 Allow() 是线程安全的
func TestTokenBucketConcurrent(t *testing.T) {
	rate := 100
	burst := 50
	limiter := NewTokenBucket(rate, burst)

	goroutines := 100
	requestsPerGoroutine := 10

	var wg sync.WaitGroup
	allowed := make(chan bool, goroutines*requestsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				allowed <- limiter.Allow()
			}
		}()
	}

	wg.Wait()
	close(allowed)

	// 统计允许的请求数
	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}

	// 初始应该至少有 burst 个请求被允许
	if count < burst {
		t.Errorf("allowed requests = %d, want at least %d", count, burst)
	}

	// 不应该超过 burst + 补充的令牌数（粗略检查）
	maxExpected := burst + rate*2 // 给予一些余量
	if count > maxExpected {
		t.Errorf("allowed requests = %d, want at most %d", count, maxExpected)
	}
}

// TestSlidingWindowWithinLimit 测试滑动窗口限流器在窗口内请求数未超限时允许通过
func TestSlidingWindowWithinLimit(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		window time.Duration
		reqs   int
	}{
		{"within_limit_5", 5, time.Second, 5},
		{"within_limit_10", 10, time.Second, 10},
		{"within_limit_1", 1, time.Second, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewSlidingWindow(tt.limit, tt.window)

			// 在限制内的请求应该全部通过
			for i := 0; i < tt.reqs; i++ {
				if !limiter.Allow() {
					t.Errorf("request %d should be allowed (limit=%d)", i+1, tt.limit)
				}
			}

			// 下一个请求应该被限流
			if limiter.Allow() {
				t.Error("expected rate limit after reaching limit")
			}
		})
	}
}

// TestSlidingWindowExceedLimit 测试滑动窗口限流器在窗口内请求数超限时拒绝
func TestSlidingWindowExceedLimit(t *testing.T) {
	limit := 5
	window := time.Second
	limiter := NewSlidingWindow(limit, window)

	// 发送 limit 个请求
	for i := 0; i < limit; i++ {
		if !limiter.Allow() {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 超出限制的请求应该被拒绝
	for i := 0; i < 3; i++ {
		if limiter.Allow() {
			t.Errorf("request %d should be rate limited", limit+i+1)
		}
	}
}

// TestSlidingWindowSliding 测试滑动窗口正确滑动，过期请求被清理
func TestSlidingWindowSliding(t *testing.T) {
	limit := 3
	window := 200 * time.Millisecond
	limiter := NewSlidingWindow(limit, window)

	// 发送 limit 个请求
	for i := 0; i < limit; i++ {
		if !limiter.Allow() {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 立即发送应该被限流
	if limiter.Allow() {
		t.Error("should be rate limited immediately after limit")
	}

	// 等待窗口过期
	time.Sleep(window + 50*time.Millisecond)

	// 窗口滑动后，应该可以再次发送请求
	for i := 0; i < limit; i++ {
		if !limiter.Allow() {
			t.Errorf("request %d should be allowed after window slides", i+1)
		}
	}
}

// TestSlidingWindowConcurrent 测试并发调用滑动窗口限流器是线程安全的
func TestSlidingWindowConcurrent(t *testing.T) {
	limit := 50
	window := time.Second
	limiter := NewSlidingWindow(limit, window)

	goroutines := 100
	requestsPerGoroutine := 5

	var wg sync.WaitGroup
	allowed := make(chan bool, goroutines*requestsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				allowed <- limiter.Allow()
			}
		}()
	}

	wg.Wait()
	close(allowed)

	// 统计允许的请求数
	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}

	// 应该恰好有 limit 个请求被允许
	if count != limit {
		t.Errorf("allowed requests = %d, want exactly %d", count, limit)
	}
}

// TestSlidingWindowWait 测试滑动窗口的 Wait 方法
func TestSlidingWindowWait(t *testing.T) {
	limit := 2
	window := 200 * time.Millisecond
	limiter := NewSlidingWindow(limit, window)

	// 消耗所有配额
	for i := 0; i < limit; i++ {
		limiter.Allow()
	}

	// Wait 应该阻塞直到窗口滑动
	ctx := context.Background()
	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait returned error: %v", err)
	}

	// 应该等待了接近 window 的时间
	if elapsed < window/2 {
		t.Errorf("Wait returned too quickly: %v", elapsed)
	}

	// Wait 之后应该可以发送请求
	if !limiter.Allow() {
		t.Error("should be allowed after Wait")
	}
}
