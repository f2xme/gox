package memadapter

import (
	"context"
	"sync"
	"sync/atomic"

	goxconfig "github.com/f2xme/gox/config"
	"github.com/f2xme/gox/queue"
)

// memQueue is an in-memory queue implementation using Go channels.
type memQueue struct {
	mu     sync.RWMutex
	cfg    Options
	topics map[string][]*subscription
	closed atomic.Bool
}

// subscription represents an active subscription to a topic.
type subscription struct {
	topic   string
	handler queue.Handler
	ch      chan *queue.Message
	done    chan struct{}
	q       *memQueue
	closed  atomic.Bool
}

// New creates a new in-memory queue with the given options.
func New(opts ...Option) queue.Queue {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &memQueue{
		cfg:    cfg,
		topics: make(map[string][]*subscription),
	}
}

// Publish sends a message to all subscribers of the specified topic.
// Returns queue.ErrClosed if the queue has been closed.
func (q *memQueue) Publish(ctx context.Context, topic string, body []byte) error {
	if q.closed.Load() {
		return queue.ErrClosed
	}

	q.mu.RLock()
	subs := q.topics[topic]
	if len(subs) == 0 {
		q.mu.RUnlock()
		return nil
	}

	// Copy body to prevent external mutation
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)

	// Copy subscription slice to avoid holding lock during send
	subsCopy := make([]*subscription, len(subs))
	copy(subsCopy, subs)
	q.mu.RUnlock()

	msg := &queue.Message{
		Topic: topic,
		Body:  bodyCopy,
	}

	for _, sub := range subsCopy {
		select {
		case sub.ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Subscribe registers a handler for the specified topic.
// Returns queue.ErrClosed if the queue has been closed.
func (q *memQueue) Subscribe(ctx context.Context, topic string, handler queue.Handler) (queue.Subscription, error) {
	if q.closed.Load() {
		return nil, queue.ErrClosed
	}

	sub := &subscription{
		topic:   topic,
		handler: handler,
		ch:      make(chan *queue.Message, q.cfg.BufferSize),
		done:    make(chan struct{}),
		q:       q,
	}

	q.mu.Lock()
	q.topics[topic] = append(q.topics[topic], sub)
	q.mu.Unlock()

	// Start consumer goroutine
	go sub.consume(ctx)

	return sub, nil
}

// consume processes messages from the subscription channel.
func (s *subscription) consume(ctx context.Context) {
	for {
		select {
		case msg, ok := <-s.ch:
			if !ok {
				return
			}
			_ = s.handler(ctx, msg)
		case <-s.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Unsubscribe stops receiving messages and removes the subscription.
func (s *subscription) Unsubscribe() error {
	if s.closed.Swap(true) {
		return nil // Already unsubscribed
	}

	close(s.done)

	s.q.mu.Lock()
	defer s.q.mu.Unlock()

	subs := s.q.topics[s.topic]
	for i, sub := range subs {
		if sub == s {
			s.q.topics[s.topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	close(s.ch)
	return nil
}

// Close stops all subscriptions and releases resources.
func (q *memQueue) Close() error {
	if q.closed.Swap(true) {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for topic, subs := range q.topics {
		for _, sub := range subs {
			if !sub.closed.Swap(true) {
				close(sub.done)
				close(sub.ch)
			}
		}
		delete(q.topics, topic)
	}

	return nil
}

// NewWithConfig creates a new in-memory queue with configuration from config.Config.
// Configuration keys:
//   - queue.mem.bufferSize (int): channel buffer size per topic (default: 64)
func NewWithConfig(cfg goxconfig.Config) queue.Queue {
	opts := []Option{}

	if bufferSize := cfg.GetInt("queue.mem.bufferSize"); bufferSize > 0 {
		opts = append(opts, WithBufferSize(bufferSize))
	}

	return New(opts...)
}
