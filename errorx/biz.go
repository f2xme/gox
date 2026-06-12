package errorx

import "errors"

// BizError 表示轻量级业务错误。
//
// 业务错误用于表达用户可预期的失败场景，例如手机号已注册、验证码错误、
// 库存不足等。这类错误不捕获堆栈信息，适合直接返回给上层做错误码识别和
// 多语言消息展示。
type BizError struct {
	// Code 错误码，用于错误识别和多语言消息查找
	Code string
	// Message 默认错误消息，作为多语言查找失败时的 fallback
	Message string
}

// NewBiz 创建业务错误。
//
// 使用当前默认语言查找已注册的错误消息；如果错误码未注册，则使用 code 作为默认消息。
func NewBiz(code string) *BizError {
	return NewBizLang(code, getDefaultLang())
}

// NewBizWithMessage 创建带显式消息的业务错误。
//
// 适用于错误消息包含动态内容，或不需要多语言注册表的场景。
func NewBizWithMessage(code, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// NewBizMessage 创建仅包含消息的业务错误。
//
// 适用于只需要向用户展示提示文案，不需要错误码识别和多语言注册表的场景。
func NewBizMessage(message string) *BizError {
	return &BizError{
		Message: message,
	}
}

// NewBizLang 创建指定语言的业务错误。
//
// 如果指定语言没有注册消息，会按注册表的 fallback 规则查找默认语言和英文消息；
// 如果错误码未注册，则使用 code 作为默认消息。
func NewBizLang(code, lang string) *BizError {
	message := code
	if msg, ok := getMessage(code, lang); ok {
		message = msg
	}
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// Localize 返回指定语言的业务错误消息。
//
// 如果指定语言没有注册消息，则返回业务错误的默认消息。
func (e *BizError) Localize(lang string) string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	if msg, ok := getMessage(e.Code, lang); ok {
		return msg
	}
	return e.Message
}

// Error 实现 error 接口，返回业务错误的默认消息。
func (e *BizError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// IsBiz 判断错误链中是否包含业务错误。
func IsBiz(err error) bool {
	if err == nil {
		return false
	}
	var e *BizError
	return errors.As(err, &e)
}

// GetBizCode 获取业务错误码。
//
// 如果错误链中不包含业务错误，返回空字符串。
func GetBizCode(err error) string {
	var e *BizError
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// LocalizeBiz 本地化业务错误消息。
//
// 如果错误链中不包含业务错误，则返回 err.Error()；如果 err 为 nil，返回空字符串。
func LocalizeBiz(err error, lang string) string {
	if err == nil {
		return ""
	}
	var e *BizError
	if errors.As(err, &e) {
		return e.Localize(lang)
	}
	return err.Error()
}
