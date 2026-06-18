package elasticsearch

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	elastic "github.com/elastic/go-elasticsearch/v8"

	"github.com/f2xme/gox/config"
)

// New 使用给定选项创建 Elasticsearch 客户端。
func New(opts ...Option) (*Client, error) {
	return NewContext(context.Background(), opts...)
}

// NewContext 使用给定选项和 context 创建 Elasticsearch 客户端。
func NewContext(ctx context.Context, opts ...Option) (*Client, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	if len(o.Addresses) == 0 && o.CloudID == "" {
		return nil, fmt.Errorf("elastic: addresses or cloud id is required")
	}

	transport := o.Transport
	if transport == nil {
		transport = &http.Transport{
			MaxIdleConnsPerHost:   o.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: o.ResponseHeaderTimeout,
			IdleConnTimeout:       o.IdleConnTimeout,
			DialContext:           (&net.Dialer{Timeout: o.DialTimeout}).DialContext,
		}
	}

	native, err := elastic.NewClient(elastic.Config{
		Addresses:    o.Addresses,
		Username:     o.Username,
		Password:     o.Password,
		CloudID:      o.CloudID,
		APIKey:       o.APIKey,
		ServiceToken: o.ServiceToken,
		MaxRetries:   o.MaxRetries,
		Transport:    transport,
	})
	if err != nil {
		return nil, fmt.Errorf("elastic: create client: %w", err)
	}

	client := &Client{client: native}
	if !o.SkipPing {
		if err := client.Ping(ctx); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// MustNew 创建 Elasticsearch 客户端，失败时退出程序。
func MustNew(opts ...Option) *Client {
	client, err := New(opts...)
	if err != nil {
		log.Fatalf("elastic: create client failed: %v", err)
	}
	return client
}

// NewWithConfig 使用 config.Config 创建 Elasticsearch 客户端。
//
// 配置键（prefix 默认为 "es"）：
//   - {prefix}.addresses ([]string): 节点地址
//   - {prefix}.apiKey (string): API Key
//   - {prefix}.username (string): Basic Auth 用户名
//   - {prefix}.password (string): Basic Auth 密码
//   - {prefix}.cloudId (string): Elastic Cloud ID
//   - {prefix}.serviceToken (string): Service Token
//   - {prefix}.maxRetries (int): 最大重试次数
//   - {prefix}.maxIdleConnsPerHost (int): 每个 Host 最大空闲连接数
//   - {prefix}.responseHeaderTimeout (duration): 响应头超时
//   - {prefix}.dialTimeout (duration): 建连超时
//   - {prefix}.idleConnTimeout (duration): 空闲连接超时
//   - {prefix}.skipPing (bool): 是否跳过连通性检查
func NewWithConfig(cfg config.Config, prefixes ...string) (*Client, error) {
	prefix := "es"
	if len(prefixes) > 0 && prefixes[0] != "" {
		prefix = prefixes[0]
	}
	key := func(suffix string) string { return prefix + "." + suffix }

	opts := make([]Option, 0, 12)
	if addresses := cfg.GetStringSlice(key("addresses")); len(addresses) > 0 {
		opts = append(opts, WithAddresses(addresses...))
	}
	if apiKey := cfg.GetString(key("apiKey")); apiKey != "" {
		opts = append(opts, WithAPIKey(apiKey))
	}
	username := cfg.GetString(key("username"))
	password := cfg.GetString(key("password"))
	if username != "" || password != "" {
		opts = append(opts, WithBasicAuth(username, password))
	}
	if cloudID := cfg.GetString(key("cloudId")); cloudID != "" {
		opts = append(opts, WithCloudID(cloudID))
	}
	if serviceToken := cfg.GetString(key("serviceToken")); serviceToken != "" {
		opts = append(opts, WithServiceToken(serviceToken))
	}
	if maxRetries := cfg.GetInt(key("maxRetries")); maxRetries > 0 {
		opts = append(opts, WithMaxRetries(maxRetries))
	}
	if maxIdle := cfg.GetInt(key("maxIdleConnsPerHost")); maxIdle > 0 {
		opts = append(opts, WithMaxIdleConnsPerHost(maxIdle))
	}
	if timeout := cfg.GetDuration(key("responseHeaderTimeout")); timeout > 0 {
		opts = append(opts, WithResponseHeaderTimeout(timeout))
	}
	if timeout := cfg.GetDuration(key("dialTimeout")); timeout > 0 {
		opts = append(opts, WithDialTimeout(timeout))
	}
	if timeout := cfg.GetDuration(key("idleConnTimeout")); timeout > 0 {
		opts = append(opts, WithIdleConnTimeout(timeout))
	}
	if cfg.GetBool(key("skipPing")) {
		opts = append(opts, WithSkipPing(true))
	}

	return New(opts...)
}

// MustNewWithConfig 使用 config.Config 创建 Elasticsearch 客户端，失败时退出程序。
func MustNewWithConfig(cfg config.Config, prefixes ...string) *Client {
	client, err := NewWithConfig(cfg, prefixes...)
	if err != nil {
		log.Fatalf("elastic: create client from config failed: %v", err)
	}
	return client
}
