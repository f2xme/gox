package rocketmq

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

// rocketmqQueue 是 RocketMQ 队列实现
type rocketmqQueue struct {
	cfg         Options
	credentials primitive.Credentials
	producer    rocketmq.Producer
	mu          sync.RWMutex
	subs        map[string]*subscription
	closed      atomic.Bool
}

// subscription 表示对主题的活动订阅
type subscription struct {
	key      string // topic:consumerGroup 复合键
	topic    string
	consumer rocketmq.PushConsumer
	q        *rocketmqQueue
	closed   atomic.Bool
}

// New 使用给定选项创建新的 RocketMQ 队列
func New(opts ...Option) (queue.Queue, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	// 创建认证凭证（复用）
	credentials := primitive.Credentials{
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}

	// Create producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(cfg.NameServers),
		producer.WithRetry(cfg.Retries),
		producer.WithGroupName(cfg.GroupName),
		producer.WithNamespace(cfg.Namespace),
		producer.WithCredentials(credentials),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start producer: %w", err)
	}

	return &rocketmqQueue{
		cfg:         cfg,
		credentials: credentials,
		producer:    p,
		subs:        make(map[string]*subscription),
	}, nil
}

// Publish 向指定主题发送消息
func (q *rocketmqQueue) Publish(ctx context.Context, topic string, body []byte) error {
	return q.PublishWithOptions(ctx, topic, body, queue.PublishOptions{})
}

// PublishWithOptions 使用额外选项发送消息
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

// Subscribe 为指定主题注册处理函数
func (q *rocketmqQueue) Subscribe(ctx context.Context, topic string, handler queue.Handler) (queue.Subscription, error) {
	return q.SubscribeWithOptions(ctx, topic, handler, queue.SubscribeOptions{
		ConsumerGroup: "DEFAULT_CONSUMER_GROUP",
		Tags:          "*",
	})
}

// SubscribeWithOptions 使用额外选项注册处理函数
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
		consumer.WithCredentials(q.credentials),
	}

	// Set consumer model
	if q.cfg.ConsumerModel == ConsumerModelBroadcasting {
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

	// 使用 topic:consumerGroup 作为复合键
	key := topic + ":" + opts.ConsumerGroup
	sub := &subscription{
		key:      key,
		topic:    topic,
		consumer: c,
		q:        q,
	}

	q.mu.Lock()
	q.subs[key] = sub
	q.mu.Unlock()

	return sub, nil
}

// Unsubscribe 停止接收消息并释放资源
func (s *subscription) Unsubscribe() error {
	if s.closed.Swap(true) {
		return nil
	}

	s.q.mu.Lock()
	delete(s.q.subs, s.key)
	s.q.mu.Unlock()

	return s.consumer.Shutdown()
}

// Close 停止所有订阅并释放资源
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
	q.subs = nil

	// Shutdown producer
	return q.producer.Shutdown()
}
