package errorx

import (
	"sync"
)

var (
	// registry 存储错误码的多语言消息
	// 结构: map[code]map[lang]message
	registry = make(map[string]map[string]string)
	// registryMu 保护 registry 的并发访问
	registryMu sync.RWMutex
	// defaultLang 错误消息的默认语言
	defaultLang = defaultEnglish
)

// Register 为特定语言注册错误码及其消息
// 此函数是并发安全的
func Register(code, lang, message string) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if registry[code] == nil {
		registry[code] = make(map[string]string)
	}
	registry[code][lang] = message
}

// SetDefaultLang 设置错误消息的默认语言
func SetDefaultLang(lang string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	defaultLang = lang
}

const defaultEnglish = "en"

// getMessage 获取指定错误码和语言的消息
func getMessage(code, lang string) (string, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	if messages, ok := registry[code]; ok {
		if msg, ok := messages[lang]; ok {
			return msg, true
		}
		if lang != defaultLang {
			if msg, ok := messages[defaultLang]; ok {
				return msg, true
			}
		}
		if lang != defaultEnglish && defaultLang != defaultEnglish {
			if msg, ok := messages[defaultEnglish]; ok {
				return msg, true
			}
		}
	}
	return "", false
}

// NewCodeWithLang 创建带有给定错误码和语言的新错误
// 如果错误码已注册，则使用注册的消息
// 否则，使用错误码作为消息
func NewCodeWithLang(code, lang string) *Error {
	message := code
	if msg, ok := getMessage(code, lang); ok {
		message = msg
	}

	return &Error{
		Code:    code,
		Message: message,
		Kind:    KindUnknown,
		Stack:   captureStack(2),
	}
}
