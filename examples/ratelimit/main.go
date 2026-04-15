package main

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/ratelimit"
)

func main() {
	fmt.Println("=== ratelimit 包使用示例 ===")

	// 示例 1: Token Bucket 限流器
	fmt.Println("\n示例 1: Token Bucket 限流器")
	fmt.Println("配置：每秒 5 个令牌，桶容量 10")

	// 创建限流器：每秒 5 个令牌，桶容量 10
	tb := ratelimit.NewTokenBucket(5, 10)

	fmt.Println("\n快速请求 15 次:")
	for i := 1; i <= 15; i++ {
		if tb.Allow() {
			fmt.Printf("  请求 %d: ✓ 通过\n", i)
		} else {
			fmt.Printf("  请求 %d: ✗ 被限流\n", i)
		}
	}

	// 等待令牌恢复
	fmt.Println("\n等待 2 秒让令牌恢复...")
	time.Sleep(2 * time.Second)

	fmt.Println("再次请求 5 次:")
	for i := 1; i <= 5; i++ {
		if tb.Allow() {
			fmt.Printf("  请求 %d: ✓ 通过\n", i)
		} else {
			fmt.Printf("  请求 %d: ✗ 被限流\n", i)
		}
	}

	// 示例 2: Sliding Window 限流器
	fmt.Println("\n示例 2: Sliding Window 限流器")
	fmt.Println("配置：每 5 秒最多 10 个请求")

	// 创建限流器：5 秒窗口，最多 10 个请求
	sw := ratelimit.NewSlidingWindow(10, 5*time.Second)

	fmt.Println("\n快速请求 12 次:")
	for i := 1; i <= 12; i++ {
		if sw.Allow() {
			fmt.Printf("  请求 %d: ✓ 通过\n", i)
		} else {
			fmt.Printf("  请求 %d: ✗ 被限流\n", i)
		}
	}

	// 示例 3: Wait 方法（阻塞等待）
	fmt.Println("\n示例 3: Wait 方法（阻塞等待）")
	tb2 := ratelimit.NewTokenBucket(2, 2)

	fmt.Println("使用 Wait 方法发送 5 个请求（会自动等待）:")
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		start := time.Now()
		if err := tb2.Wait(ctx); err != nil {
			fmt.Printf("  请求 %d: 错误 %v\n", i, err)
		} else {
			elapsed := time.Since(start)
			fmt.Printf("  请求 %d: ✓ 通过（等待 %v）\n", i, elapsed.Round(time.Millisecond))
		}
	}

	// 示例 4: Reserve 方法（预约）
	fmt.Println("\n示例 4: Reserve 方法（预约）")
	tb3 := ratelimit.NewTokenBucket(1, 1)

	for i := 1; i <= 3; i++ {
		r := tb3.Reserve()
		if !r.OK() {
			fmt.Printf("  请求 %d: ✗ 预约失败\n", i)
			continue
		}

		delay := r.Delay()
		if delay > 0 {
			fmt.Printf("  请求 %d: 需要等待 %v\n", i, delay.Round(time.Millisecond))
			time.Sleep(delay)
		}
		fmt.Printf("  请求 %d: ✓ 执行\n", i)
	}

	// 示例 5: 带超时的 Wait
	fmt.Println("\n示例 5: 带超时的 Wait")
	tb4 := ratelimit.NewTokenBucket(1, 1)

	// 先消耗所有令牌
	tb4.Allow()

	// 尝试等待，但设置 100ms 超时
	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	fmt.Println("尝试获取令牌（超时 100ms）:")
	if err := tb4.Wait(ctx2); err != nil {
		fmt.Printf("  ✗ 超时: %v\n", err)
	} else {
		fmt.Println("  ✓ 获取成功")
	}

	fmt.Println("\n=== 示例结束 ===")
	fmt.Println("\n提示：")
	fmt.Println("- Token Bucket: 允许突发流量，适合 API 限流")
	fmt.Println("- Sliding Window: 严格限制时间窗口内的请求数")
	fmt.Println("- Allow(): 非阻塞，立即返回")
	fmt.Println("- Wait(): 阻塞等待，直到可以通过")
	fmt.Println("- Reserve(): 预约令牌，返回需要等待的时间")
}
