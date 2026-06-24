package memory

import (
	"log"
)

// New 创建新的内存短信适配器。
func New(opts ...Option) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Client{
		options: options,
		records: make([]SentMessage, 0),
	}, nil
}

// MustNew 创建新的内存短信适配器，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Client {
	client, err := New(opts...)
	if err != nil {
		log.Fatalf("memory sms: create client failed: %v", err)
	}
	return client
}
