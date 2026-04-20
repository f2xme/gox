package rocketmq

import "time"

// 消费模式常量
const (
	ConsumerModelClustering   = "clustering"   // 集群消费模式
	ConsumerModelBroadcasting = "broadcasting" // 广播消费模式（RocketMQ 5.x 不支持，仅作保留）
)

// Options 定义 RocketMQ 队列的配置选项
type Options struct {
	// Endpoint RocketMQ 5.x Proxy 地址（格式：host:port）
	Endpoint string
	// AccessKey 认证访问密钥（可选）
	AccessKey string
	// SecretKey 认证密钥（可选）
	SecretKey string
	// Namespace 消息隔离命名空间（可选）
	Namespace string
	// Retries 发送失败重试次数（保留字段，新 SDK 内部处理重试）
	Retries int
	// SendTimeout 发送消息超时时间
	SendTimeout time.Duration
	// ConsumerModel 消费模式（集群或广播）
	ConsumerModel string
	// Topics 预热的 topic 列表，启动时会预先拉取路由信息。
	// 注意：至少需要指定一个 topic，否则 SDK 无法完成初始化握手，导致 New/NewContext 永久阻塞。
	Topics []string
}

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		Endpoint:      "127.0.0.1:8081",
		Retries:       2,
		SendTimeout:   3 * time.Second,
		ConsumerModel: ConsumerModelClustering,
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithEndpoint 设置 RocketMQ 5.x Proxy 地址
//
// 示例：
//
//	rocketmq.New(rocketmq.WithEndpoint("localhost:8081"))
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithCredentials 设置认证的访问密钥和密钥
//
// 示例：
//
//	rocketmq.New(rocketmq.WithCredentials("myAccessKey", "mySecretKey"))
func WithCredentials(accessKey, secretKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
		o.SecretKey = secretKey
	}
}

// WithNamespace 设置消息隔离的命名空间
//
// 示例：
//
//	rocketmq.New(rocketmq.WithNamespace("dev"))
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithRetries 设置发送失败的重试次数
//
// 示例：
//
//	rocketmq.New(rocketmq.WithRetries(3))
func WithRetries(retries int) Option {
	return func(o *Options) {
		if retries < 0 {
			retries = 0
		}
		o.Retries = retries
	}
}

// WithSendTimeout 设置发送消息的超时时间
//
// 示例：
//
//	rocketmq.New(rocketmq.WithSendTimeout(5 * time.Second))
func WithSendTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		if timeout <= 0 {
			timeout = 3 * time.Second
		}
		o.SendTimeout = timeout
	}
}

// WithTopics 设置启动时预热的 topic 列表，可加速首次发送
//
// 示例：
//
//	rocketmq.New(rocketmq.WithTopics("orders", "payments"))
func WithTopics(topics ...string) Option {
	return func(o *Options) {
		o.Topics = topics
	}
}
//
// 示例：
//
//	rocketmq.New(rocketmq.WithConsumerModel("clustering"))
func WithConsumerModel(model string) Option {
	return func(o *Options) {
		o.ConsumerModel = model
	}
}
