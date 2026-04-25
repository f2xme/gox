// Package generator 定义验证码生成器接口。
package generator

// Generator 定义验证码生成器接口。
// 实现此接口可以创建自定义的验证码生成器。
type Generator interface {
	// Generate 生成验证码内容和答案。
	// data: base64 编码的验证码数据（图片或音频）
	// answer: 验证码答案（用于验证用户输入）
	Generate() (data string, answer string, err error)

	// Type 返回生成器类型标识。
	// 用于日志、监控和调试。
	Type() string
}
