// Package volcengine provides a placeholder for a future Volcengine SMS
// adapter.
//
// 当前版本不会创建真实的火山引擎短信客户端，也不会发送短信。
// New、NewWithConfig 和 NewWithOptions 在配置校验通过后返回
// ErrNotImplemented，避免调用方误以为该适配器已经可用于生产。
package volcengine
