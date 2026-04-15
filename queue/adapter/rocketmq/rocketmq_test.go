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
			name: "with custom name servers",
			opts: []Option{
				WithNameServers([]string{"localhost:9876"}),
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
			name: "with group name",
			opts: []Option{
				WithGroupName("test-group"),
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
				WithConsumerModel("broadcasting"),
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
	t.Run("WithNameServers", func(t *testing.T) {
		opts := defaultOptions()
		WithNameServers([]string{"server1:9876", "server2:9876"})(&opts)
		if len(opts.NameServers) != 2 {
			t.Errorf("expected 2 name servers, got %d", len(opts.NameServers))
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

	t.Run("WithGroupName", func(t *testing.T) {
		opts := defaultOptions()
		WithGroupName("test-group")(&opts)
		if opts.GroupName != "test-group" {
			t.Errorf("group name not set correctly")
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
		WithConsumerModel("broadcasting")(&opts)
		if opts.ConsumerModel != "broadcasting" {
			t.Errorf("consumer model not set correctly")
		}
	})
}

// TestDefaultOptions tests default options.
func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if len(opts.NameServers) != 1 || opts.NameServers[0] != "127.0.0.1:9876" {
		t.Errorf("unexpected default name servers: %v", opts.NameServers)
	}

	if opts.GroupName != "DEFAULT_PRODUCER_GROUP" {
		t.Errorf("unexpected default group name: %s", opts.GroupName)
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

// Note: Integration tests require a running RocketMQ server.
// The following tests are examples and will be skipped in CI.

func TestIntegration_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	q, err := New(
		WithNameServers([]string{"localhost:9876"}),
		WithGroupName("test-producer"),
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
		WithNameServers([]string{"localhost:9876"}),
		WithGroupName("test-producer"),
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
