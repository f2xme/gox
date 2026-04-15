// Package queue 提供统一的消息队列抽象层。
// 支持内存队列和分布式队列的类型安全操作。
package queue

import (
	"context"
	"time"
)

// Message 表示队列中的一条消息
type Message struct {
	// ID 消息的唯一标识符
	ID string
	// Topic 消息所属的主题/频道
	Topic string
	// Body 消息负载
	Body []byte
	// Tags 可选的消息标签，用于过滤（RocketMQ 等使用）
	Tags string
	// Keys 可选的消息键，用于索引和搜索
	Keys []string
	// Properties 可选的键值对，用于自定义元数据
	Properties map[string]string
	// DelayLevel 延迟消息的延迟级别（0 = 不延迟）
	DelayLevel int
	// BornTimestamp 消息创建时间
	BornTimestamp time.Time
}

// Handler 消息处理回调函数
// 返回 nil 表示确认消息，返回 error 表示拒绝消息
type Handler func(ctx context.Context, msg *Message) error

// PublishOptions 发布消息的可选参数
type PublishOptions struct {
	// Tags 消息标签，用于过滤（例如 "TagA||TagB"）
	Tags string
	// Keys 消息键，用于索引
	Keys []string
	// Properties 自定义元数据
	Properties map[string]string
	// DelayLevel 延迟投递级别（0 = 立即投递）
	DelayLevel int
}

// SubscribeOptions 订阅主题的可选参数
type SubscribeOptions struct {
	// ConsumerGroup 消费者组名称（分布式队列必需）
	ConsumerGroup string
	// Tags 过滤消息的标签（例如 "TagA||TagB"，"*" 表示全部）
	Tags string
	// MaxConcurrency 限制并发消息处理数量（0 = 不限制）
	MaxConcurrency int
	// AutoCommit 启用自动消息确认（默认：true）
	AutoCommit bool
}

// Publisher 定义发布操作接口
type Publisher interface {
	// Publish 发送消息到指定主题
	Publish(ctx context.Context, topic string, body []byte) error
	// PublishWithOptions 使用额外选项发送消息
	PublishWithOptions(ctx context.Context, topic string, body []byte, opts PublishOptions) error
}

// Subscriber 定义订阅操作接口
type Subscriber interface {
	// Subscribe 为指定主题注册处理函数
	// 每当主题收到消息时，处理函数会被调用
	// 返回一个 Subscription，可用于取消订阅
	Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
	// SubscribeWithOptions 使用额外选项注册处理函数
	SubscribeWithOptions(ctx context.Context, topic string, handler Handler, opts SubscribeOptions) (Subscription, error)
}

// Subscription 表示一个可以关闭的活动订阅
type Subscription interface {
	// Unsubscribe 停止接收消息并释放资源
	Unsubscribe() error
}

// Queue 组合了 Publisher 和 Subscriber 接口
type Queue interface {
	Publisher
	Subscriber
}

// Closer 为队列实现提供资源清理
type Closer interface {
	// Close 释放队列持有的所有资源
	Close() error
}
