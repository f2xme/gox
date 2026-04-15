package rocketmqadapter

import "time"

// Options holds the configuration for RocketMQ queue.
type Options struct {
	// NameServers is the list of RocketMQ name server addresses.
	NameServers []string
	// AccessKey for authentication (optional).
	AccessKey string
	// SecretKey for authentication (optional).
	SecretKey string
	// Namespace for message isolation (optional).
	Namespace string
	// GroupName is the default producer group name.
	GroupName string
	// Retries is the number of retry attempts for failed sends.
	Retries int
	// SendTimeout is the timeout for sending messages.
	SendTimeout time.Duration
	// ConsumerModel specifies clustering or broadcasting mode.
	ConsumerModel string
}

// defaultOptions returns default configuration.
func defaultOptions() Options {
	return Options{
		NameServers: []string{"127.0.0.1:9876"},
		GroupName:   "DEFAULT_PRODUCER_GROUP",
		Retries:     2,
		SendTimeout: 3 * time.Second,
		ConsumerModel: "clustering",
	}
}

// Option is a function that modifies Options.
type Option func(*Options)

// WithNameServers sets the RocketMQ name server addresses.
//
// Example:
//
//	rocketmq.New(rocketmq.WithNameServers([]string{"localhost:9876"}))
func WithNameServers(addrs []string) Option {
	return func(o *Options) {
		o.NameServers = addrs
	}
}

// WithCredentials sets the access key and secret key for authentication.
//
// Example:
//
//	rocketmq.New(rocketmq.WithCredentials("myAccessKey", "mySecretKey"))
func WithCredentials(accessKey, secretKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
		o.SecretKey = secretKey
	}
}

// WithNamespace sets the namespace for message isolation.
//
// Example:
//
//	rocketmq.New(rocketmq.WithNamespace("dev"))
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithGroupName sets the producer group name.
//
// Example:
//
//	rocketmq.New(rocketmq.WithGroupName("my-producer-group"))
func WithGroupName(groupName string) Option {
	return func(o *Options) {
		o.GroupName = groupName
	}
}

// WithRetries sets the number of retry attempts for failed sends.
//
// Example:
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

// WithSendTimeout sets the timeout for sending messages.
//
// Example:
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

// WithConsumerModel sets the consumer model (clustering or broadcasting).
//
// Example:
//
//	rocketmq.New(rocketmq.WithConsumerModel("broadcasting"))
func WithConsumerModel(model string) Option {
	return func(o *Options) {
		o.ConsumerModel = model
	}
}
