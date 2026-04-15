package ratelimit_test

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/ratelimit"
)

// Example_tokenBucket demonstrates basic usage of token bucket rate limiter.
func Example_tokenBucket() {
	// Create a limiter: 10 requests per second, burst of 5
	limiter := ratelimit.NewTokenBucket(10, 5)

	// First 5 requests should be allowed (burst)
	for i := 0; i < 5; i++ {
		if limiter.Allow() {
			fmt.Println("Request allowed")
		}
	}

	// 6th request should be rate limited
	if !limiter.Allow() {
		fmt.Println("Rate limited")
	}

	// Output:
	// Request allowed
	// Request allowed
	// Request allowed
	// Request allowed
	// Request allowed
	// Rate limited
}

// Example_slidingWindow demonstrates basic usage of sliding window rate limiter.
func Example_slidingWindow() {
	// Create a limiter: 3 requests per 100ms window
	limiter := ratelimit.NewSlidingWindow(3, 100*time.Millisecond)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if limiter.Allow() {
			fmt.Println("Request allowed")
		}
	}

	// 4th request should be rate limited
	if !limiter.Allow() {
		fmt.Println("Rate limited")
	}

	// Output:
	// Request allowed
	// Request allowed
	// Request allowed
	// Rate limited
}

// Example_wait demonstrates using Wait() to block until a request can proceed.
func Example_wait() {
	limiter := ratelimit.NewTokenBucket(2, 1)

	// Consume the initial token
	limiter.Allow()

	// Wait for the next token (will block briefly)
	ctx := context.Background()
	if err := limiter.Wait(ctx); err != nil {
		fmt.Println("Wait failed:", err)
		return
	}

	fmt.Println("Request allowed after waiting")

	// Output:
	// Request allowed after waiting
}

// Example_httpMiddleware demonstrates integrating rate limiter with HTTP middleware.
func Example_httpMiddleware() {
	// This is a conceptual example (not runnable)
	limiter := ratelimit.NewTokenBucket(100, 10)

	// Middleware function
	_ = func(next func()) func() {
		return func() {
			if !limiter.Allow() {
				fmt.Println("HTTP 429: Rate limit exceeded")
				return
			}
			next()
		}
	}

	fmt.Println("Middleware created")

	// Output:
	// Middleware created
}

// Example_reserve demonstrates using Reserve() to check wait time without blocking.
func Example_reserve() {
	limiter := ratelimit.NewTokenBucket(10, 2)

	// Consume all tokens
	limiter.Allow()
	limiter.Allow()

	// Reserve the next token
	reservation := limiter.Reserve()

	if reservation.OK() {
		delay := reservation.Delay()
		if delay > 0 {
			fmt.Println("Need to wait before proceeding")
		} else {
			fmt.Println("Can proceed immediately")
		}
	}

	// Output:
	// Need to wait before proceeding
}

// Example_contextCancellation demonstrates handling context cancellation with Wait().
func Example_contextCancellation() {
	limiter := ratelimit.NewTokenBucket(1, 1)

	// Consume the token
	limiter.Allow()

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait will return error when context is cancelled
	if err := limiter.Wait(ctx); err != nil {
		fmt.Println("Wait cancelled")
	}

	// Output:
	// Wait cancelled
}
