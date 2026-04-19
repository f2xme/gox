package rocketmq

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/queue"
)

// TestNew tests creating a new RocketMQ queue.
func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "default options",
			opts: []Option{},
		},
		{
			name: "with custom endpoint",
			opts: []Option{
				WithEndpoint("localhost:8081"),
			},
		},
		{
			name: "with credentials",
			opts: []Option{
				WithCredentials("accessKey", "secretKey"),
			},
		},
		{
			name: "with namespace",
			opts: []Option{
				WithNamespace("test"),
			},
		},
		{
			name: "with retries",
			opts: []Option{
				WithRetries(3),
			},
		},
		{
			name: "with send timeout",
			opts: []Option{
				WithSendTimeout(5 * time.Second),
			},
		},
		{
			name: "with consumer model",
			opts: []Option{
				WithConsumerModel("clustering"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(tt.opts...)
			// Note: New() may succeed even without a RocketMQ server
			// because the producer is created but not yet connected.
			// Actual connection happens on first Publish/Subscribe.
			if err != nil {
				t.Logf("New() returned error (expected without RocketMQ server): %v", err)
			}
			if q != nil {
				if closer, ok := q.(queue.Closer); ok {
					_ = closer.Close()
				}
			}
		})
	}
}

// TestOptions tests option functions.
func TestOptions(t *testing.T) {
	t.Run("WithEndpoint", func(t *testing.T) {
		opts := defaultOptions()
		WithEndpoint("server1:8081")(&opts)
		if opts.Endpoint != "server1:8081" {
			t.Errorf("expected endpoint 'server1:8081', got %s", opts.Endpoint)
		}
	})

	t.Run("WithCredentials", func(t *testing.T) {
		opts := defaultOptions()
		WithCredentials("key", "secret")(&opts)
		if opts.AccessKey != "key" || opts.SecretKey != "secret" {
			t.Errorf("credentials not set correctly")
		}
	})

	t.Run("WithNamespace", func(t *testing.T) {
		opts := defaultOptions()
		WithNamespace("test-ns")(&opts)
		if opts.Namespace != "test-ns" {
			t.Errorf("namespace not set correctly")
		}
	})

	t.Run("WithRetries", func(t *testing.T) {
		opts := defaultOptions()
		WithRetries(5)(&opts)
		if opts.Retries != 5 {
			t.Errorf("retries not set correctly")
		}
	})

	t.Run("WithRetries negative", func(t *testing.T) {
		opts := defaultOptions()
		WithRetries(-1)(&opts)
		if opts.Retries != 0 {
			t.Errorf("expected retries to be 0, got %d", opts.Retries)
		}
	})

	t.Run("WithSendTimeout", func(t *testing.T) {
		opts := defaultOptions()
		WithSendTimeout(10 * time.Second)(&opts)
		if opts.SendTimeout != 10*time.Second {
			t.Errorf("send timeout not set correctly")
		}
	})

	t.Run("WithSendTimeout zero", func(t *testing.T) {
		opts := defaultOptions()
		WithSendTimeout(0)(&opts)
		if opts.SendTimeout != 3*time.Second {
			t.Errorf("expected default timeout, got %v", opts.SendTimeout)
		}
	})

	t.Run("WithConsumerModel", func(t *testing.T) {
		opts := defaultOptions()
		WithConsumerModel("clustering")(&opts)
		if opts.ConsumerModel != "clustering" {
			t.Errorf("consumer model not set correctly")
		}
	})
}

// TestDefaultOptions tests default options.
func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if opts.Endpoint != "127.0.0.1:8081" {
		t.Errorf("unexpected default endpoint: %v", opts.Endpoint)
	}

	if opts.Retries != 2 {
		t.Errorf("unexpected default retries: %d", opts.Retries)
	}

	if opts.SendTimeout != 3*time.Second {
		t.Errorf("unexpected default send timeout: %v", opts.SendTimeout)
	}

	if opts.ConsumerModel != "clustering" {
		t.Errorf("unexpected default consumer model: %s", opts.ConsumerModel)
	}
}

// 注意：集成测试需要运行中的 RocketMQ 服务，以下测试在 CI 中会被跳过。

func TestIntegration_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	q, err := New(
		WithEndpoint("localhost:8081"),
	)
	if err != nil {
		t.Skipf("skipping test, RocketMQ not available: %v", err)
	}
	defer q.(queue.Closer).Close()

	// Subscribe
	received := make(chan *queue.Message, 1)
	sub, err := q.SubscribeWithOptions(ctx, "test-topic", func(ctx context.Context, msg *queue.Message) error {
		received <- msg
		return nil
	}, queue.SubscribeOptions{
		ConsumerGroup: "test-consumer",
		Tags:          "*",
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Wait for consumer to be ready
	time.Sleep(2 * time.Second)

	// Publish
	testBody := []byte("test message")
	err = q.PublishWithOptions(ctx, "test-topic", testBody, queue.PublishOptions{
		Tags: "test",
		Keys: []string{"test-key"},
	})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for message
	select {
	case msg := <-received:
		if string(msg.Body) != string(testBody) {
			t.Errorf("expected body %s, got %s", testBody, msg.Body)
		}
		if msg.Tags != "test" {
			t.Errorf("expected tags 'test', got %s", msg.Tags)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestIntegration_DelayMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	q, err := New(
		WithEndpoint("localhost:8081"),
	)
	if err != nil {
		t.Skipf("skipping test, RocketMQ not available: %v", err)
	}
	defer q.(queue.Closer).Close()

	// Subscribe
	received := make(chan time.Time, 1)
	sub, err := q.SubscribeWithOptions(ctx, "test-delay-topic", func(ctx context.Context, msg *queue.Message) error {
		received <- time.Now()
		return nil
	}, queue.SubscribeOptions{
		ConsumerGroup: "test-delay-consumer",
		Tags:          "*",
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Wait for consumer to be ready
	time.Sleep(2 * time.Second)

	// Publish delayed message (level 2 = 5 seconds)
	start := time.Now()
	err = q.PublishWithOptions(ctx, "test-delay-topic", []byte("delayed"), queue.PublishOptions{
		DelayLevel: 2,
	})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for message
	select {
	case receivedTime := <-received:
		delay := receivedTime.Sub(start)
		if delay < 4*time.Second || delay > 7*time.Second {
			t.Errorf("expected delay around 5s, got %v", delay)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timeout waiting for delayed message")
	}
}
