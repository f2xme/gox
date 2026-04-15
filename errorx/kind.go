package errorx

// Kind 表示错误的类别
type Kind int

const (
	// KindUnknown 表示未知错误类型
	KindUnknown Kind = iota
	// KindValidation 表示验证错误
	KindValidation
	// KindNotFound 表示资源未找到错误
	KindNotFound
	// KindConflict 表示冲突错误（例如重复键）
	KindConflict
	// KindUnauthorized 表示认证错误
	KindUnauthorized
	// KindForbidden 表示授权错误
	KindForbidden
	// KindInternal 表示内部服务器错误
	KindInternal
	// KindTimeout 表示超时错误
	KindTimeout
	// KindRetryable 表示可重试的错误
	KindRetryable
)

func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindValidation:
		return "Validation"
	case KindNotFound:
		return "NotFound"
	case KindConflict:
		return "Conflict"
	case KindUnauthorized:
		return "Unauthorized"
	case KindForbidden:
		return "Forbidden"
	case KindInternal:
		return "Internal"
	case KindTimeout:
		return "Timeout"
	case KindRetryable:
		return "Retryable"
	default:
		return "Unknown"
	}
}

// IsRetryable 如果错误类别可重试则返回 true
func (k Kind) IsRetryable() bool {
	return k == KindRetryable || k == KindTimeout
}
