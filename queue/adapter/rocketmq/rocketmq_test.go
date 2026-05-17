package rocketmq

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/queue"
)

// TestNew requires topics before touching the RocketMQ SDK.
func TestNew(t *testing.T) {
	q, err := New(WithEndpoint("localhost:8081"))
	if err == nil {
		if closer, ok := q.(queue.Closer); ok {
			_ = closer.Close()
		}
		t.Fatal("expected error when topics are not configured")
	}
	if !strings.Contains(err.Error(), "WithTopics") {
		t.Fatalf("expected WithTopics error, got %v", err)
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

	t.Run("WithTopics", func(t *testing.T) {
		opts := defaultOptions()
		WithTopics("orders", "payments")(&opts)
		if len(opts.Topics) != 2 || opts.Topics[0] != "orders" || opts.Topics[1] != "payments" {
			t.Errorf("topics not set correctly: %v", opts.Topics)
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

	if len(opts.Topics) != 0 {
		t.Errorf("unexpected default topics: %v", opts.Topics)
	}
}

func skipRocketMQIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if os.Getenv("GOX_ROCKETMQ_INTEGRATION") != "1" {
		t.Skip("set GOX_ROCKETMQ_INTEGRATION=1 to run RocketMQ integration tests")
	}
}

func TestIntegration_PublishSubscribe(t *testing.T) {
	skipRocketMQIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q, err := NewContext(ctx,
		WithEndpoint("localhost:8081"),
		WithTopics("test-topic"),
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
	skipRocketMQIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q, err := NewContext(ctx,
		WithEndpoint("localhost:8081"),
		WithTopics("test-delay-topic"),
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
