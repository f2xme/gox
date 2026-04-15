package ratelimit

import (
	"context"
	"time"
)

// Limiter 定义限流器接口
type Limiter interface {
	// Allow 检查事件是否允许通过（非阻塞）
	// 返回 true 表示允许，false 表示被限流
	Allow() bool

	// Wait 阻塞直到事件可以通过或 context 被取消
	// 如果 context 被取消则返回错误
	Wait(ctx context.Context) error

	// Reserve 预留一个事件并返回 Reservation
	Reserve() Reservation
}

// Reservation 表示限流器中的一个预留事件
type Reservation interface {
	// OK 返回预留是否有效
	OK() bool

	// Delay 返回事件可以通过前需要等待的时间
	Delay() time.Duration

	// Cancel 取消预留
	Cancel()
}
