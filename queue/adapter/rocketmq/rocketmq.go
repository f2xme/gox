package rocketmqadapter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/f2xme/gox/queue"
)

// rocketmqQueue is a RocketMQ queue implementation.
type rocketmqQueue struct {
	cfg      Options
	producer rocketmq.Producer
	mu       sync.RWMutex
	subs     map[string]*subscription
	closed   atomic.Bool
}

// subscription represents an active subscription to a topic.
type subscription struct {
	topic    string
	consumer rocketmq.PushConsumer
	q        *rocketmqQueue
	closed   atomic.Bool
}

// New creates a new RocketMQ queue with the given options.
func New(opts ...Option) (queue.Queue, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Create producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(cfg.NameServers),
		producer.WithRetry(cfg.Retries),
		producer.WithGroupName(cfg.GroupName),
		producer.WithNamespace(cfg.Namespace),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start producer: %w", err)
	}

	return &rocketmqQueue{
		cfg:      cfg,
		producer: p,
		subs:     make(map[string]*subscription),
	}, nil
}

// Publish sends a message to the specified topic.
func (q *rocketmqQueue) Publish(ctx context.Context, topic string, body []byte) error {
	return q.PublishWithOptions(ctx, topic, body, queue.PublishOptions{})
}

// PublishWithOptions sends a message with additional options.
func (q *rocketmqQueue) PublishWithOptions(ctx context.Context, topic string, body []byte, opts queue.PublishOptions) error {
	if q.closed.Load() {
		return queue.ErrClosed
	}

	msg := primitive.NewMessage(topic, body)

	// Set tags
	if opts.Tags != "" {
		msg.WithTag(opts.Tags)
	}

	// Set keys
	if len(opts.Keys) > 0 {
		msg.WithKeys(opts.Keys)
	}

	// Set properties
	if len(opts.Properties) > 0 {
		for k, v := range opts.Properties {
			msg.WithProperty(k, v)
		}
	}

	// Set delay level
	if opts.DelayLevel > 0 {
		msg.WithDelayTimeLevel(opts.DelayLevel)
	}

	// Send message with timeout
	sendCtx, cancel := context.WithTimeout(ctx, q.cfg.SendTimeout)
	defer cancel()

	_, err := q.producer.SendSync(sendCtx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Subscribe registers a handler for the specified topic.
func (q *rocketmqQueue) Subscribe(ctx context.Context, topic string, handler queue.Handler) (queue.Subscription, error) {
	return q.SubscribeWithOptions(ctx, topic, handler, queue.SubscribeOptions{
		ConsumerGroup: "DEFAULT_CONSUMER_GROUP",
		Tags:          "*",
	})
}

// SubscribeWithOptions registers a handler with additional options.
func (q *rocketmqQueue) SubscribeWithOptions(ctx context.Context, topic string, handler queue.Handler, opts queue.SubscribeOptions) (queue.Subscription, error) {
	if q.closed.Load() {
		return nil, queue.ErrClosed
	}

	if opts.ConsumerGroup == "" {
		return nil, fmt.Errorf("consumer group is required")
	}

	// Create consumer options
	consumerOpts := []consumer.Option{
		consumer.WithNameServer(q.cfg.NameServers),
		consumer.WithGroupName(opts.ConsumerGroup),
		consumer.WithNamespace(q.cfg.Namespace),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: q.cfg.AccessKey,
			SecretKey: q.cfg.SecretKey,
		}),
	}

	// Set consumer model
	if q.cfg.ConsumerModel == "broadcasting" {
		consumerOpts = append(consumerOpts, consumer.WithConsumerModel(consumer.BroadCasting))
	} else {
		consumerOpts = append(consumerOpts, consumer.WithConsumerModel(consumer.Clustering))
	}

	// Set max concurrency
	if opts.MaxConcurrency > 0 {
		consumerOpts = append(consumerOpts, consumer.WithConsumeMessageBatchMaxSize(opts.MaxConcurrency))
	}

	// Create consumer
	c, err := rocketmq.NewPushConsumer(consumerOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Subscribe to topic
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: opts.Tags,
	}

	if opts.Tags == "" {
		selector.Expression = "*"
	}

	err = c.Subscribe(topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			queueMsg := &queue.Message{
				ID:            msg.MsgId,
				Topic:         msg.Topic,
				Body:          msg.Body,
				Tags:          msg.GetTags(),
				Keys:          []string{msg.GetKeys()},
				Properties:    msg.GetProperties(),
				BornTimestamp: time.Unix(msg.BornTimestamp/1000, 0),
			}

			if err := handler(ctx, queueMsg); err != nil {
				// Return consume later to retry
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	// Start consumer
	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	sub := &subscription{
		topic:    topic,
		consumer: c,
		q:        q,
	}

	q.mu.Lock()
	q.subs[topic] = sub
	q.mu.Unlock()

	return sub, nil
}

// Unsubscribe stops receiving messages and releases resources.
func (s *subscription) Unsubscribe() error {
	if s.closed.Swap(true) {
		return nil
	}

	s.q.mu.Lock()
	delete(s.q.subs, s.topic)
	s.q.mu.Unlock()

	return s.consumer.Shutdown()
}

// Close stops all subscriptions and releases resources.
func (q *rocketmqQueue) Close() error {
	if q.closed.Swap(true) {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Close all subscriptions
	for _, sub := range q.subs {
		if !sub.closed.Swap(true) {
			_ = sub.consumer.Shutdown()
		}
	}
	q.subs = make(map[string]*subscription)

	// Shutdown producer
	return q.producer.Shutdown()
}
