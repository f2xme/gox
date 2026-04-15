package rocketmq

import "time"

// 消费模式常量
const (
	ConsumerModelClustering  = "clustering"  // 集群消费模式
	ConsumerModelBroadcasting = "broadcasting" // 广播消费模式
)

// Options 定义 RocketMQ 队列的配置选项
type Options struct {
	// NameServers RocketMQ name server 地址列表
	NameServers []string
	// AccessKey 认证访问密钥（可选）
	AccessKey string
	// SecretKey 认证密钥（可选）
	SecretKey string
	// Namespace 消息隔离命名空间（可选）
	Namespace string
	// GroupName 默认生产者组名称
	GroupName string
	// Retries 发送失败重试次数
	Retries int
	// SendTimeout 发送消息超时时间
	SendTimeout time.Duration
	// ConsumerModel 消费模式（集群或广播）
	ConsumerModel string
}

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		NameServers:   []string{"127.0.0.1:9876"},
		GroupName:     "DEFAULT_PRODUCER_GROUP",
		Retries:       2,
		SendTimeout:   3 * time.Second,
		ConsumerModel: ConsumerModelClustering,
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithNameServers 设置 RocketMQ name server 地址
//
// 示例：
//
//	rocketmq.New(rocketmq.WithNameServers([]string{"localhost:9876"}))
func WithNameServers(addrs []string) Option {
	return func(o *Options) {
		o.NameServers = addrs
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

// WithGroupName 设置生产者组名称
//
// 示例：
//
//	rocketmq.New(rocketmq.WithGroupName("my-producer-group"))
func WithGroupName(groupName string) Option {
	return func(o *Options) {
		o.GroupName = groupName
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

// WithConsumerModel 设置消费模式（集群或广播）
//
// 示例：
//
//	rocketmq.New(rocketmq.WithConsumerModel("broadcasting"))
func WithConsumerModel(model string) Option {
	return func(o *Options) {
		o.ConsumerModel = model
	}
}
