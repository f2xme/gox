// Package queue provides a unified abstraction layer for message queue operations.
// It supports both in-memory and distributed queues with type-safe operations.
package queue

import "context"

// Message represents a message in the queue.
type Message struct {
	// ID is the unique identifier of the message.
	ID string
	// Topic is the topic/channel the message belongs to.
	Topic string
	// Body is the message payload.
	Body []byte
}

// Handler is a callback function for processing messages.
// Return nil to acknowledge the message, or an error to nack it.
type Handler func(ctx context.Context, msg *Message) error

// Publisher defines the publish operations interface.
type Publisher interface {
	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, body []byte) error
}

// Subscriber defines the subscribe operations interface.
type Subscriber interface {
	// Subscribe registers a handler for the specified topic.
	// The handler is called for each message received on the topic.
	// Returns a Subscription that can be used to unsubscribe.
	Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
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
