// Package base64 提供基于 base64Captcha 库的验证码生成器。
//
// 该生成器实现 captcha.Generator 接口，负责生成验证码展示数据和答案。
// 返回的数据可作为图片或音频的 base64 内容交给调用方展示。
//
// # 功能特性
//
//   - 数字验证码：生成纯数字图片验证码
//   - 字符验证码：生成字母数字混合图片验证码
//   - 算术验证码：生成需要计算答案的图片验证码
//   - 音频验证码：生成语音朗读数字的音频验证码
//
// # 快速开始
//
// 创建 base64 生成器：
//
//	gen, err := base64.New(
//		base64.WithType(base64.TypeDigit),
//		base64.WithLength(6),
//		base64.WithSize(300, 100),
//	)
//	if err != nil {
//		return err
//	}
//
//	data, err := gen.Generate(context.Background())
//	if err != nil {
//		return err
//	}
//
// # 注意事项
//
//   - New 会校验尺寸、长度、噪点数量和验证码类型
//   - 图片验证码返回 PNG 图片的 base64 字符串
//   - 音频验证码返回 WAV 音频的 base64 字符串
//   - answer 只应保存在服务端，不应返回给客户端
package base64
