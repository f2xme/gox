package email

import (
	"log"

	"github.com/f2xme/gox/config"
	"gopkg.in/gomail.v2"
)

// Client 邮件客户端
type Client struct {
	name   string
	dialer *gomail.Dialer
}

// New 创建邮件客户端
//
// 参数：
//   - host: SMTP 服务器地址
//   - port: SMTP 端口
//   - username: 发件人邮箱地址
//   - password: 邮箱密码或授权码
//   - opts: 可选配置项
//
// 示例：
//
//	client, err := email.New("smtp.gmail.com", 587, "user@gmail.com", "password")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 使用可选配置
//	client, err := email.New("smtp.gmail.com", 587, "user@gmail.com", "password",
//		email.WithName("My App"),
//		email.WithSSL(true),
//	)
func New(host string, port int, username, password string, opts ...Option) (*Client, error) {
	options := defaultOptions()
	options.Host = host
	options.Port = port
	options.Username = username
	options.Password = password

	for _, opt := range opts {
		opt(&options)
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}

	dialer := gomail.NewDialer(options.Host, options.Port, options.Username, options.Password)
	dialer.SSL = options.SSL

	return &Client{
		name:   options.Name,
		dialer: dialer,
	}, nil
}

// NewWithOptions 使用 Options 模式创建邮件客户端
//
// 这是推荐的创建方式，提供更灵活的配置选项。
//
// 示例：
//
//	client, err := email.NewWithOptions(
//		email.WithHost("smtp.gmail.com"),
//		email.WithPort(587),
//		email.WithUsername("user@gmail.com"),
//		email.WithPassword("password"),
//		email.WithName("My App"),
//		email.WithSSL(false),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
func NewWithOptions(opts ...Option) (*Client, error) {
	options := defaultOptions()

	for _, opt := range opts {
		opt(&options)
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}

	dialer := gomail.NewDialer(options.Host, options.Port, options.Username, options.Password)
	dialer.SSL = options.SSL

	return &Client{
		name:   options.Name,
		dialer: dialer,
	}, nil
}

// NewWithConfig 使用 config.Config 中的配置创建邮件客户端
// 配置键：
//   - email.host (string): SMTP 服务器地址（必需）
//   - email.port (int): SMTP 端口（必需）
//   - email.username (string): 发件人邮箱地址（必需）
//   - email.password (string): 邮箱密码或授权码（必需）
//   - email.name (string): 发件人名称
//   - email.ssl (bool): 是否使用 SSL
func NewWithConfig(cfg config.Config) (*Client, error) {
	return New(
		cfg.GetString("email.host"),
		cfg.GetInt("email.port"),
		cfg.GetString("email.username"),
		cfg.GetString("email.password"),
		WithName(cfg.GetString("email.name")),
		WithSSL(cfg.GetBool("email.ssl")),
	)
}

// MustNewWithConfig 使用配置创建邮件客户端，失败时终止程序
func MustNewWithConfig(cfg config.Config) *Client {
	client, err := NewWithConfig(cfg)
	if err != nil {
		log.Fatalf("email: failed to create client from config: %v", err)
	}
	return client
}
