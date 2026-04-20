package rocketmq

import (
	"context"
	"fmt"
	"log"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/queue"
)

// buildCredentials 根据 Options 构建认证凭证，无需认证时返回 nil
func buildCredentials(cfg Options) *credentials.SessionCredentials {
	if cfg.AccessKey == "" && cfg.SecretKey == "" {
		return nil
	}
	return &credentials.SessionCredentials{
		AccessKey:    cfg.AccessKey,
		AccessSecret: cfg.SecretKey,
	}
}

// buildConfig 根据 Options 构建 SDK Config，consumerGroup 为空时用于 Producer
func buildConfig(cfg Options, consumerGroup string) *rmq.Config {
	return &rmq.Config{
		Endpoint:      cfg.Endpoint,
		NameSpace:     cfg.Namespace,
		ConsumerGroup: consumerGroup,
		Credentials:   buildCredentials(cfg),
	}
}

// New 使用给定选项创建新的 RocketMQ 队列
func New(opts ...Option) (queue.Queue, error) {
	return NewContext(context.Background(), opts...)
}

// NewContext 使用给定选项和 context 创建新的 RocketMQ 队列，context 可用于控制启动超时
func NewContext(ctx context.Context, opts ...Option) (queue.Queue, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	if len(cfg.Topics) == 0 {
		return nil, fmt.Errorf("rocketmq: at least one topic is required, use WithTopics to specify")
	}

	p, err := rmq.NewProducer(buildConfig(cfg, ""), rmq.WithTopics(cfg.Topics...))
	if err != nil {
		return nil, fmt.Errorf("rocketmq: failed to create producer: %w", err)
	}

	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		ch <- result{err: p.Start()}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return nil, fmt.Errorf("rocketmq: failed to start producer: %w", r.err)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("rocketmq: producer start timeout: %w", ctx.Err())
	}

	return &rocketmqQueue{
		cfg:      cfg,
		producer: p,
		subs:     make(map[string]*subscription),
	}, nil
}

// MustNew 创建一个新的 RocketMQ 队列，出错时终止程序。
func MustNew(opts ...Option) queue.Queue {
	q, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return q
}

// NewWithConfig 使用 config.Config 中的配置创建一个新的 RocketMQ 队列。
// 配置键：
//   - queue.rocketmq.endpoint (string): Proxy 地址（默认："127.0.0.1:8081"）
//   - queue.rocketmq.accessKey (string): 认证访问密钥
//   - queue.rocketmq.secretKey (string): 认证密钥
//   - queue.rocketmq.namespace (string): 命名空间
//   - queue.rocketmq.retries (int): 发送失败重试次数（默认：2）
//   - queue.rocketmq.sendTimeout (duration): 发送消息超时时间（默认：3s）
//   - queue.rocketmq.consumerModel (string): 消费模式（默认："clustering"）
//
// 示例：
//
//	q, err := rocketmq.NewWithConfig(cfg)
func NewWithConfig(cfg config.Config) (queue.Queue, error) {
	var opts []Option

	if endpoint := cfg.GetString("queue.rocketmq.endpoint"); endpoint != "" {
		opts = append(opts, WithEndpoint(endpoint))
	}

	accessKey := cfg.GetString("queue.rocketmq.accessKey")
	secretKey := cfg.GetString("queue.rocketmq.secretKey")
	if accessKey != "" || secretKey != "" {
		opts = append(opts, WithCredentials(accessKey, secretKey))
	}

	if namespace := cfg.GetString("queue.rocketmq.namespace"); namespace != "" {
		opts = append(opts, WithNamespace(namespace))
	}

	if retries := cfg.GetInt("queue.rocketmq.retries"); retries >= 0 {
		opts = append(opts, WithRetries(retries))
	}

	if timeout := cfg.GetDuration("queue.rocketmq.sendTimeout"); timeout > 0 {
		opts = append(opts, WithSendTimeout(timeout))
	}

	if consumerModel := cfg.GetString("queue.rocketmq.consumerModel"); consumerModel != "" {
		opts = append(opts, WithConsumerModel(consumerModel))
	}

	return New(opts...)
}

// MustNewWithConfig 使用配置创建一个新的 RocketMQ 队列，出错时终止程序。
func MustNewWithConfig(cfg config.Config) queue.Queue {
	q, err := NewWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return q
}
