// Package base64 提供基于 base64Captcha 库的验证码生成器。
//
// 支持多种验证码类型：
//   - TypeDigit: 纯数字验证码
//   - TypeString: 字母数字混合验证码
//   - TypeMath: 算术表达式验证码
//   - TypeAudio: 音频验证码
//
// 示例：
//
//	gen := base64.New(
//		base64.WithType(base64.TypeDigit),
//		base64.WithLength(6),
//		base64.WithSize(300, 100),
//	)
//	data, answer, err := gen.Generate()
package base64
