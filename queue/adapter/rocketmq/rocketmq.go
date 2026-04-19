package rocketmq

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/f2xme/gox/queue"
)

// delayLevels 映射 RocketMQ 延迟等级（1-18）到时间间隔
var delayLevels = []time.Duration{
	0,                // 占位（level 0 表示不延迟）
	time.Second,      // 1
	5 * time.Second,  // 2
	10 * time.Second, // 3
	30 * time.Second, // 4
	time.Minute,      // 5
	2 * time.Minute,  // 6
	3 * time.Minute,  // 7
	4 * time.Minute,  // 8
	5 * time.Minute,  // 9
	6 * time.Minute,  // 10
	7 * time.Minute,  // 11
	8 * time.Minute,  // 12
	9 * time.Minute,  // 13
	10 * time.Minute, // 14
	20 * time.Minute, // 15
	30 * time.Minute, // 16
	time.Hour,        // 17
	2 * time.Hour,    // 18
}

// rocketmqQueue 是 RocketMQ 队列实现
type rocketmqQueue struct {
	cfg      Options
	producer rmq.Producer
	mu       sync.RWMutex
	subs     map[string]*subscription
	closed   atomic.Bool
}

// subscription 表示对主题的活动订阅
type subscription struct {
	topic         string
	consumerGroup string
	consumer      rmq.PushConsumer
	q             *rocketmqQueue
	closed        atomic.Bool
}

// Publish 向指定主题发送消息
func (q *rocketmqQueue) Publish(ctx context.Context, topic string, body []byte) error {
	return q.PublishWithOptions(ctx, topic, body, queue.PublishOptions{})
}

// PublishWithOptions 使用额外选项发送消息
func (q *rocketmqQueue) PublishWithOptions(ctx context.Context, topic string, body []byte, opts queue.PublishOptions) error {
	_, err := q.PublishAndGetResult(ctx, topic, body, opts)
	return err
}

// PublishAndGetResult 发送消息并返回包含消息 ID 的结果，实现 queue.AdvancedPublisher
func (q *rocketmqQueue) PublishAndGetResult(ctx context.Context, topic string, body []byte, opts queue.PublishOptions) (*queue.SendResult, error) {
	if q.closed.Load() {
		return nil, queue.ErrClosed
	}

	msg := &rmq.Message{
		Topic: topic,
		Body:  body,
	}

	if opts.Tags != "" {
		msg.SetTag(opts.Tags)
	}

	if len(opts.Keys) > 0 {
		msg.SetKeys(opts.Keys...)
	}

	for k, v := range opts.Properties {
		msg.AddProperty(k, v)
	}

	// 将延迟等级转换为绝对投递时间
	if opts.DelayLevel >= 1 && opts.DelayLevel <= 18 {
		msg.SetDelayTimestamp(time.Now().Add(delayLevels[opts.DelayLevel]))
	}

	sendCtx, cancel := context.WithTimeout(ctx, q.cfg.SendTimeout)
	defer cancel()

	receipts, err := q.producer.Send(sendCtx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	if len(receipts) == 0 {
		return nil, fmt.Errorf("send message returned empty receipts")
	}

	return &queue.SendResult{MessageID: receipts[0].MessageID}, nil
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
		return nil, fmt.Errorf("rocketmq: consumer group is required")
	}

	// 构建过滤表达式
	expression := opts.Tags
	if expression == "" {
		expression = "*"
	}
	filterExpr := rmq.NewFilterExpression(expression)

	// 创建 PushConsumer，通过 MessageListener 注册消息处理回调
	c, err := rmq.NewPushConsumer(
		buildConfig(q.cfg, opts.ConsumerGroup),
		rmq.WithPushSubscriptionExpressions(map[string]*rmq.FilterExpression{
			topic: filterExpr,
		}),
		rmq.WithPushMessageListener(&rmq.FuncMessageListener{
			Consume: func(mv *rmq.MessageView) rmq.ConsumerResult {
				tag := ""
				if t := mv.GetTag(); t != nil {
					tag = *t
				}
				bornTime := time.Time{}
				if bt := mv.GetBornTimestamp(); bt != nil {
					bornTime = *bt
				}
				queueMsg := &queue.Message{
					ID:            mv.GetMessageId(),
					Topic:         mv.GetTopic(),
					Body:          mv.GetBody(),
					Tags:          tag,
					Keys:          mv.GetKeys(),
					Properties:    mv.GetProperties(),
					BornTimestamp: bornTime,
				}
				if err := handler(ctx, queueMsg); err != nil {
					return rmq.FAILURE
				}
				return rmq.SUCCESS
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	sub := &subscription{
		topic:         topic,
		consumerGroup: opts.ConsumerGroup,
		consumer:      c,
		q:             q,
	}

	q.mu.Lock()
	q.subs[topic+":"+opts.ConsumerGroup] = sub
	q.mu.Unlock()

	return sub, nil
}

// Unsubscribe 停止接收消息并释放资源
func (s *subscription) Unsubscribe() error {
	if s.closed.Swap(true) {
		return nil
	}

	s.q.mu.Lock()
	delete(s.q.subs, s.topic+":"+s.consumerGroup)
	s.q.mu.Unlock()

	return s.consumer.GracefulStop()
}

// Close 停止所有订阅并释放资源
func (q *rocketmqQueue) Close() error {
	if q.closed.Swap(true) {
		return nil
	}

	q.mu.Lock()
	subs := q.subs
	q.subs = nil
	q.mu.Unlock()

	for _, sub := range subs {
		if !sub.closed.Swap(true) {
			_ = sub.consumer.GracefulStop()
		}
	}

	return q.producer.GracefulStop()
}
