package mem

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/f2xme/gox/queue"
)

// memQueue 是使用 Go channel 实现的内存队列
type memQueue struct {
	mu     sync.RWMutex
	cfg    Options
	topics map[string][]*subscription
	closed atomic.Bool
}

// subscription 表示对主题的活动订阅
type subscription struct {
	topic   string
	handler queue.Handler
	ch      chan *queue.Message
	done    chan struct{}
	q       *memQueue
	closed  atomic.Bool
}

// New 使用给定选项创建新的内存队列
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

// Publish 向指定主题的所有订阅者发送消息
// 如果队列已关闭，返回 queue.ErrClosed
func (q *memQueue) Publish(ctx context.Context, topic string, body []byte) error {
	return q.PublishWithOptions(ctx, topic, body, queue.PublishOptions{})
}

// PublishWithOptions 使用额外选项发送消息
func (q *memQueue) PublishWithOptions(ctx context.Context, topic string, body []byte, opts queue.PublishOptions) error {
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
		Topic:      topic,
		Body:       bodyCopy,
		Tags:       opts.Tags,
		Keys:       opts.Keys,
		Properties: opts.Properties,
		DelayLevel: opts.DelayLevel,
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

// Subscribe 为指定主题注册处理函数
// 如果队列已关闭，返回 queue.ErrClosed
func (q *memQueue) Subscribe(ctx context.Context, topic string, handler queue.Handler) (queue.Subscription, error) {
	return q.SubscribeWithOptions(ctx, topic, handler, queue.SubscribeOptions{})
}

// SubscribeWithOptions 使用额外选项注册处理函数
func (q *memQueue) SubscribeWithOptions(ctx context.Context, topic string, handler queue.Handler, opts queue.SubscribeOptions) (queue.Subscription, error) {
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

// consume 从订阅通道处理消息
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

// Unsubscribe 停止接收消息并移除订阅
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

// Close 停止所有订阅并释放资源
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

