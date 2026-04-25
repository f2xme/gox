package captcha

import "context"

// Challenge 表示一次验证码挑战。
type Challenge struct {
	// ID 验证码唯一标识
	ID string
	// Data 验证码展示数据，通常是图片或音频的 base64 内容
	Data string
	// Type 验证码生成器类型
	Type string
}

// ChallengeData 表示生成器生成的验证码数据。
type ChallengeData struct {
	// Data 验证码展示数据，通常是图片或音频的 base64 内容
	Data string
	// Answer 验证码答案，只应保存在服务端
	Answer string
}

// Generator 定义验证码生成器接口。
// 实现此接口可以创建自定义的验证码生成器。
type Generator interface {
	// Generate 生成验证码内容和答案。
	Generate(ctx context.Context) (ChallengeData, error)

	// Type 返回生成器类型标识。
	// 用于日志、监控和调试。
	Type() string
}
