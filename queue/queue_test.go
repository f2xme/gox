package queue_test

import (
	"context"
	"sync"
	"testing"

	"github.com/f2xme/gox/queue"
)

// mockQueue is a minimal in-process queue for testing the interface contract.
type mockQueue struct {
	mu     sync.RWMutex
	subs   map[string][]queue.Handler
	closed bool
}

func newMockQueue() *mockQueue {
	return &mockQueue{
		subs: make(map[string][]queue.Handler),
	}
}

func (q *mockQueue) Publish(ctx context.Context, topic string, body []byte) error {
	if q.closed {
		return queue.ErrClosed
	}

	q.mu.RLock()
	handlers := q.subs[topic]
	q.mu.RUnlock()

	msg := &queue.Message{Topic: topic, Body: body}
	for _, h := range handlers {
		if err := h(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

type mockSub struct {
	q     *mockQueue
	topic string
	index int
}

func (s *mockSub) Unsubscribe() error {
	s.q.mu.Lock()
	defer s.q.mu.Unlock()

	handlers := s.q.subs[s.topic]
	if s.index >= 0 && s.index < len(handlers) {
		s.q.subs[s.topic] = append(handlers[:s.index], handlers[s.index+1:]...)
	}
	return nil
}

func (q *mockQueue) Subscribe(_ context.Context, topic string, handler queue.Handler) (queue.Subscription, error) {
	if q.closed {
		return nil, queue.ErrClosed
	}

	q.mu.Lock()
	index := len(q.subs[topic])
	q.subs[topic] = append(q.subs[topic], handler)
	q.mu.Unlock()

	return &mockSub{q: q, topic: topic, index: index}, nil
}

func (q *mockQueue) Close() error {
	q.closed = true
	return nil
}

func TestQueueInterface(t *testing.T) {
	// Verify mockQueue satisfies Queue and Closer interfaces
	var _ queue.Queue = newMockQueue()
	var _ queue.Closer = newMockQueue()
}

func TestQueuePublishSubscribe(t *testing.T) {
	q := newMockQueue()
	ctx := context.Background()

	var received []byte
	_, err := q.Subscribe(ctx, "test", func(ctx context.Context, msg *queue.Message) error {
		received = msg.Body
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	err = q.Publish(ctx, "test", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if string(received) != "hello" {
		t.Errorf("received %q, want %q", received, "hello")
	}
}

func TestQueueClosedErrors(t *testing.T) {
	q := newMockQueue()
	ctx := context.Background()

	q.Close()

	err := q.Publish(ctx, "test", []byte("msg"))
	if err != queue.ErrClosed {
		t.Errorf("Publish after close returned %v, want %v", err, queue.ErrClosed)
	}

	_, err = q.Subscribe(ctx, "test", func(ctx context.Context, msg *queue.Message) error {
		return nil
	})
	if err != queue.ErrClosed {
		t.Errorf("Subscribe after close returned %v, want %v", err, queue.ErrClosed)
	}
}
