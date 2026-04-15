package email

import "errors"

// Options 定义邮件客户端配置选项
type Options struct {
	// Host 是 SMTP 服务器地址
	Host string
	// Port 是 SMTP 端口
	Port int
	// Username 是发件人邮箱地址
	Username string
	// Password 是邮箱密码或授权码
	Password string
	// Name 是发件人名称
	Name string
	// SSL 指示是否使用 SSL
	SSL bool
}

// defaultOptions 返回默认配置
func defaultOptions() Options {
	return Options{
		SSL: false,
	}
}

// Validate 验证配置选项的有效性
func (o *Options) Validate() error {
	if o.Host == "" {
		return errors.New("email: host is required")
	}
	if o.Port <= 0 {
		return errors.New("email: port must be positive")
	}
	if o.Username == "" {
		return errors.New("email: username is required")
	}
	if o.Password == "" {
		return errors.New("email: password is required")
	}
	return nil
}

// Option 定义配置选项函数
type Option func(*Options)

// WithHost 设置 SMTP 服务器地址
//
// 示例：
//
//	email.NewWithOptions(email.WithHost("smtp.gmail.com"))
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort 设置 SMTP 端口
//
// 示例：
//
//	email.NewWithOptions(email.WithPort(587))
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithUsername 设置发件人邮箱地址
//
// 示例：
//
//	email.NewWithOptions(email.WithUsername("sender@example.com"))
func WithUsername(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// WithPassword 设置邮箱密码或授权码
//
// 示例：
//
//	email.NewWithOptions(email.WithPassword("your-password"))
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithName 设置发件人名称
//
// 示例：
//
//	email.NewWithOptions(email.WithName("System Notification"))
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithSSL 设置是否使用 SSL
//
// 示例：
//
//	email.NewWithOptions(email.WithSSL(true))
func WithSSL(ssl bool) Option {
	return func(o *Options) {
		o.SSL = ssl
	}
}
