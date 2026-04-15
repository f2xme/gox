/*
Package ratelimit 提供多种限流算法实现。

ratelimit 包提供统一的限流接口，支持多种限流策略。
所有限流器都实现了 Limiter 接口，可以互换使用。

# 功能特性

  - Token Bucket（令牌桶）：允许突发流量，平滑速率控制
  - Sliding Window（滑动窗口）：严格限制时间窗口内的请求数
  - 统一的 Limiter 接口：所有算法可互换使用
  - 线程安全：支持多 goroutine 并发使用
  - 多种使用模式：Allow（非阻塞）、Wait（阻塞）、Reserve（预留）

# 快速开始

基本使用：

	package main

	import (
		"context"
		"fmt"
		"time"

		"github.com/f2xme/gox/ratelimit"
	)

	func main() {
		// 创建令牌桶限流器：每秒 10 个请求，允许突发 5 个
		limiter := ratelimit.NewTokenBucket(10, 5)

		// 非阻塞检查
		if limiter.Allow() {
			fmt.Println("请求被允许")
		} else {
			fmt.Println("请求被限流")
		}

		// 阻塞等待
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := limiter.Wait(ctx); err != nil {
			fmt.Println("等待超时:", err)
		} else {
			fmt.Println("请求被允许")
		}
	}

# 支持的算法

## Token Bucket - 令牌桶算法

允许突发流量，同时保持平均速率：

	limiter := ratelimit.NewTokenBucket(10, 5)
	if limiter.Allow() {
		// 请求被允许
	}

特性：
  - 允许突发流量（burst）
  - 平滑的速率控制
  - 适合 API 限流
  - 基于 golang.org/x/time/rate

参数说明：
  - rate: 每秒允许的请求数
  - burst: 最大突发请求数

## Sliding Window - 滑动窗口算法

在固定时间窗口内严格限制请求数：

	limiter := ratelimit.NewSlidingWindow(100, time.Minute)
	if limiter.Allow() {
		// 请求被允许
	}

特性：
  - 严格的速率限制
  - 精确的时间窗口
  - 自动清理过期记录
  - 适合防止突发攻击

参数说明：
  - limit: 时间窗口内允许的最大请求数
  - window: 时间窗口大小

# 使用示例

## 基本用法

非阻塞检查，立即返回：

	limiter := ratelimit.NewTokenBucket(10, 5)

	if limiter.Allow() {
		// 处理请求
		handleRequest()
	} else {
		// 返回 429 Too Many Requests
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	}

## 阻塞等待

阻塞直到请求可以通过：

	limiter := ratelimit.NewTokenBucket(10, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := limiter.Wait(ctx); err != nil {
		// 超时或 context 被取消
		return err
	}

	// 请求被允许，继续处理
	handleRequest()

## 预留检查

检查需要等待的时间，不阻塞：

	limiter := ratelimit.NewTokenBucket(10, 5)

	reservation := limiter.Reserve()
	if !reservation.OK() {
		// 无法预留
		return
	}

	delay := reservation.Delay()
	if delay > 0 {
		// 需要等待
		time.Sleep(delay)
	}

	// 处理请求
	handleRequest()

## HTTP 中间件

集成到 HTTP 服务器：

	func RateLimitMiddleware(limiter ratelimit.Limiter) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !limiter.Allow() {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
			})
		}
	}

	// 使用中间件
	limiter := ratelimit.NewTokenBucket(100, 10)
	handler := RateLimitMiddleware(limiter)(http.DefaultServeMux)
	http.ListenAndServe(":8080", handler)

## 按用户限流

为每个用户创建独立的限流器：

	type UserRateLimiter struct {
		limiters sync.Map // map[string]ratelimit.Limiter
		rate     int
		burst    int
	}

	func (u *UserRateLimiter) GetLimiter(userID string) ratelimit.Limiter {
		if limiter, ok := u.limiters.Load(userID); ok {
			return limiter.(ratelimit.Limiter)
		}

		limiter := ratelimit.NewTokenBucket(u.rate, u.burst)
		u.limiters.Store(userID, limiter)
		return limiter
	}

	func (u *UserRateLimiter) Allow(userID string) bool {
		return u.GetLimiter(userID).Allow()
	}

## 按 IP 限流

基于客户端 IP 地址限流：

	type IPRateLimiter struct {
		limiters sync.Map
		limit    int
		window   time.Duration
	}

	func (i *IPRateLimiter) Allow(ip string) bool {
		limiter, _ := i.limiters.LoadOrStore(ip,
			ratelimit.NewSlidingWindow(i.limit, i.window))
		return limiter.(ratelimit.Limiter).Allow()
	}

	// 使用
	ipLimiter := &IPRateLimiter{limit: 100, window: time.Minute}

	func handler(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !ipLimiter.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		// 处理请求
	}

## 多级限流

组合多个限流器实现多级限流：

	type MultiLimiter struct {
		limiters []ratelimit.Limiter
	}

	func (m *MultiLimiter) Allow() bool {
		for _, limiter := range m.limiters {
			if !limiter.Allow() {
				return false
			}
		}
		return true
	}

	// 使用：同时限制每秒和每分钟的请求数
	limiter := &MultiLimiter{
		limiters: []ratelimit.Limiter{
			ratelimit.NewTokenBucket(10, 5),           // 每秒 10 个
			ratelimit.NewSlidingWindow(100, time.Minute), // 每分钟 100 个
		},
	}

# 算法选择指南

| 场景 | 推荐算法 | 原因 |
|------|---------|------|
| API 限流 | Token Bucket | 允许短时突发，用户体验好 |
| 防止 DDoS | Sliding Window | 严格限制，防止突发攻击 |
| 资源保护 | Token Bucket | 平滑控制，避免资源耗尽 |
| 流量整形 | Token Bucket | 支持突发，平均速率稳定 |
| 精确计数 | Sliding Window | 精确统计时间窗口内请求数 |

# 最佳实践

## 1. 合理设置 rate 和 burst

	// API 限流：允许一定突发
	limiter := ratelimit.NewTokenBucket(100, 20) // 每秒 100 个，突发 20 个

	// 严格限流：burst 设置为 1
	limiter := ratelimit.NewTokenBucket(10, 1) // 严格每秒 10 个

	// 滑动窗口：根据业务需求设置
	limiter := ratelimit.NewSlidingWindow(1000, time.Minute) // 每分钟 1000 个

## 2. 使用 context 控制超时

	// 推荐：设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := limiter.Wait(ctx); err != nil {
		// 处理超时
		return err
	}

	// 不推荐：无限等待
	limiter.Wait(context.Background())

## 3. 为不同的端点设置不同的限流策略

	// 读操作：较高的限流
	readLimiter := ratelimit.NewTokenBucket(1000, 100)

	// 写操作：较低的限流
	writeLimiter := ratelimit.NewTokenBucket(100, 10)

	// 敏感操作：非常严格的限流
	authLimiter := ratelimit.NewSlidingWindow(5, time.Minute)

## 4. 返回友好的错误信息

	if !limiter.Allow() {
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Rate limit exceeded. Please try again later.",
			http.StatusTooManyRequests)
		return
	}

## 5. 使用 Reserve() 实现优雅降级

	reservation := limiter.Reserve()
	if !reservation.OK() {
		// 无法预留，拒绝请求
		return errors.New("rate limit exceeded")
	}

	delay := reservation.Delay()
	if delay > maxAcceptableDelay {
		// 等待时间太长，取消预留
		reservation.Cancel()
		return errors.New("rate limit exceeded")
	}

	// 等待时间可接受
	time.Sleep(delay)
	handleRequest()

## 6. 定期清理限流器缓存

对于按用户或 IP 限流的场景，定期清理不活跃的限流器：

	type UserRateLimiter struct {
		limiters sync.Map
		lastSeen sync.Map
	}

	func (u *UserRateLimiter) Cleanup() {
		now := time.Now()
		u.lastSeen.Range(func(key, value interface{}) bool {
			if now.Sub(value.(time.Time)) > time.Hour {
				u.limiters.Delete(key)
				u.lastSeen.Delete(key)
			}
			return true
		})
	}

	// 定期清理
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup()
		}
	}()

# 性能考虑

## Token Bucket

  - 基于 golang.org/x/time/rate，性能优秀
  - 内存占用小（每个限流器约 100 字节）
  - 适合高并发场景
  - 无需定期清理

## Sliding Window

  - 需要存储时间戳列表
  - 内存占用与 limit 成正比
  - 需要定期清理过期时间戳
  - 适合中等并发场景

## 性能对比

	Token Bucket:    ~10M ops/sec
	Sliding Window:  ~1M ops/sec

# 线程安全

所有限流器实现都是线程安全的，可以在多个 goroutine 中并发使用。

# 常见问题

## Q: Token Bucket 和 Sliding Window 有什么区别？

A: Token Bucket 允许突发流量，适合 API 限流；Sliding Window 严格限制时间窗口内的请求数，适合防止攻击。

## Q: 如何实现分布式限流？

A: 本包提供的是单机限流。分布式限流需要使用 Redis 等外部存储，可以参考 httpx/middleware/ratelimit 包。

## Q: 限流器会自动清理吗？

A: Token Bucket 不需要清理；Sliding Window 会在每次 Allow() 时自动清理过期时间戳。

## Q: 如何选择 burst 大小？

A: 通常设置为 rate 的 10-20%。例如 rate=100，burst=10-20。

## Q: Wait() 会一直阻塞吗？

A: 不会。使用 context.WithTimeout() 可以设置最大等待时间。

# 相关包

  - httpx/middleware/ratelimit: HTTP 限流中间件
  - cache: 缓存抽象层，可用于分布式限流
*/
package ratelimit
