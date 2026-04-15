// Package queue provides a unified abstraction layer for message queue operations.
// It supports both in-memory and distributed queues with type-safe operations.
package queue

import (
	"context"
	"time"
)

// Message represents a message in the queue.
type Message struct {
	// ID is the unique identifier of the message.
	ID string
	// Topic is the topic/channel the message belongs to.
	Topic string
	// Body is the message payload.
	Body []byte
	// Tags are optional message tags for filtering (used by RocketMQ, etc).
	Tags string
	// Keys are optional message keys for indexing and searching.
	Keys []string
	// Properties are optional key-value pairs for custom metadata.
	Properties map[string]string
	// DelayLevel specifies the delay level for delayed messages (0 = no delay).
	DelayLevel int
	// BornTimestamp is when the message was created.
	BornTimestamp time.Time
}

// Handler is a callback function for processing messages.
// Return nil to acknowledge the message, or an error to nack it.
type Handler func(ctx context.Context, msg *Message) error

// PublishOptions contains optional parameters for publishing messages.
type PublishOptions struct {
	// Tags for message filtering (e.g., "TagA||TagB").
	Tags string
	// Keys for message indexing.
	Keys []string
	// Properties for custom metadata.
	Properties map[string]string
	// DelayLevel for delayed delivery (0 = immediate).
	DelayLevel int
}

// SubscribeOptions contains optional parameters for subscribing to topics.
type SubscribeOptions struct {
	// ConsumerGroup is the consumer group name (required for distributed queues).
	ConsumerGroup string
	// Tags for filtering messages (e.g., "TagA||TagB", "*" for all).
	Tags string
	// MaxConcurrency limits concurrent message processing (0 = unlimited).
	MaxConcurrency int
	// AutoCommit enables automatic message acknowledgment (default: true).
	AutoCommit bool
}

// Publisher defines the publish operations interface.
type Publisher interface {
	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, body []byte) error
	// PublishWithOptions sends a message with additional options.
	PublishWithOptions(ctx context.Context, topic string, body []byte, opts PublishOptions) error
}

// Subscriber defines the subscribe operations interface.
type Subscriber interface {
	// Subscribe registers a handler for the specified topic.
	// The handler is called for each message received on the topic.
	// Returns a Subscription that can be used to unsubscribe.
	Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
	// SubscribeWithOptions registers a handler with additional options.
	SubscribeWithOptions(ctx context.Context, topic string, handler Handler, opts SubscribeOptions) (Subscription, error)
}

// Subscription represents an active subscription that can be closed.
type Subscription interface {
	// Unsubscribe stops receiving messages and releases resources.
	Unsubscribe() error
}

// Queue combines Publisher and Subscriber interfaces.
type Queue interface {
	Publisher
	Subscriber
}

// Closer provides resource cleanup for queue implementations.
type Closer interface {
	// Close releases any resources held by the queue.
	Close() error
}
