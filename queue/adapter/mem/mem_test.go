package mem

import (
	"context"
	"sync"
	"testing"

	"github.com/f2xme/gox/queue"
)

func TestMemQueuePublishSubscribe(t *testing.T) {
	q := New()
	defer q.(queue.Closer).Close()

	ctx := context.Background()
	topic := "test-topic"
	body := []byte("hello")

	var received []byte
	var wg sync.WaitGroup
	wg.Add(1)

	_, err := q.Subscribe(ctx, topic, func(ctx context.Context, msg *queue.Message) error {
		received = msg.Body
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	err = q.Publish(ctx, topic, body)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	wg.Wait()

	if string(received) != string(body) {
		t.Errorf("received %q, want %q", received, body)
	}
}

func TestMemQueueMultipleSubscribers(t *testing.T) {
	q := New()
	defer q.(queue.Closer).Close()

	ctx := context.Background()
	topic := "broadcast"
	body := []byte("broadcast-msg")

	var mu sync.Mutex
	var count int
	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		_, err := q.Subscribe(ctx, topic, func(ctx context.Context, msg *queue.Message) error {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe %d failed: %v", i, err)
		}
	}

	err := q.Publish(ctx, topic, body)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	wg.Wait()

	if count != 3 {
		t.Errorf("received count = %d, want 3", count)
	}
}

func TestMemQueueUnsubscribe(t *testing.T) {
	q := New()
	defer q.(queue.Closer).Close()

	ctx := context.Background()
	topic := "unsub-topic"

	var mu sync.Mutex
	var count int
	var wg sync.WaitGroup
	var once sync.Once

	wg.Add(1)
	sub, err := q.Subscribe(ctx, topic, func(ctx context.Context, msg *queue.Message) error {
		mu.Lock()
		count++
		mu.Unlock()
		once.Do(wg.Done)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish first message
	err = q.Publish(ctx, topic, []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for delivery
	wg.Wait()

	// Unsubscribe
	err = sub.Unsubscribe()
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	// Publish second message (should not be received)
	err = q.Publish(ctx, topic, []byte("msg2"))
	if err != nil {
		t.Fatalf("Publish after unsubscribe failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("received count = %d, want 1", count)
	}
}

func TestMemQueueClose(t *testing.T) {
	q := New()

	ctx := context.Background()
	topic := "close-topic"

	_, err := q.Subscribe(ctx, topic, func(ctx context.Context, msg *queue.Message) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Close the queue
	err = q.(queue.Closer).Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Publish after close should return ErrClosed
	err = q.Publish(ctx, topic, []byte("msg"))
	if err != queue.ErrClosed {
		t.Errorf("Publish after close returned %v, want %v", err, queue.ErrClosed)
	}

	// Subscribe after close should return ErrClosed
	_, err = q.Subscribe(ctx, topic, func(ctx context.Context, msg *queue.Message) error {
		return nil
	})
	if err != queue.ErrClosed {
		t.Errorf("Subscribe after close returned %v, want %v", err, queue.ErrClosed)
	}
}

func TestMemQueueNoSubscribers(t *testing.T) {
	q := New()
	defer q.(queue.Closer).Close()

	ctx := context.Background()

	// Publish to topic with no subscribers should not error
	err := q.Publish(ctx, "empty-topic", []byte("msg"))
	if err != nil {
		t.Fatalf("Publish to empty topic failed: %v", err)
	}
}

func TestMemQueueTopicIsolation(t *testing.T) {
	q := New()
	defer q.(queue.Closer).Close()

	ctx := context.Background()

	var received string
	var wg sync.WaitGroup
	wg.Add(1)

	_, err := q.Subscribe(ctx, "topic-a", func(ctx context.Context, msg *queue.Message) error {
		received = string(msg.Body)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish to different topic
	err = q.Publish(ctx, "topic-b", []byte("wrong"))
	if err != nil {
		t.Fatalf("Publish to topic-b failed: %v", err)
	}

	// Publish to subscribed topic
	err = q.Publish(ctx, "topic-a", []byte("correct"))
	if err != nil {
		t.Fatalf("Publish to topic-a failed: %v", err)
	}

	wg.Wait()

	if received != "correct" {
		t.Errorf("received %q, want %q", received, "correct")
	}
}
