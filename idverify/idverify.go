package idverify

import (
	"context"
	"strings"
	"time"
)

// 提供方编码，写入业务审计日志时使用。
const (
	ProviderMock   = "mock"
	ProviderBaidu  = "baidu"
	ProviderAliyun = "aliyun"
)

// 通用业务错误码（适配器可将厂商码映射到此，便于业务统一处理）。
const (
	// CodeNameMismatch 姓名与证件号不一致。
	CodeNameMismatch = "name_mismatch"
	// CodeIDInvalid 证件号无效或权威库查无。
	CodeIDInvalid = "id_invalid"
)

// Request 二要素核验请求。
type Request struct {
	// Name 真实姓名。
	Name string
	// IDNumber 身份证号（通常为 15/18 位居民身份证或港澳台居住证号码）。
	IDNumber string
}

// Normalize 去除首尾空白，并将身份证号转为大写（兼容末位 x/X）。
func (r Request) Normalize() Request {
	return Request{
		Name:     strings.TrimSpace(r.Name),
		IDNumber: strings.ToUpper(strings.TrimSpace(r.IDNumber)),
	}
}

// Valid 检查请求是否具备基本字段。
func (r Request) Valid() bool {
	n := r.Normalize()
	return n.Name != "" && n.IDNumber != ""
}

// Result 二要素核验业务结果。
//
// 当 error == nil 时：
//   - Matched == true 表示一致；
//   - Matched == false 表示业务不一致（姓名/证件不匹配等），应查看 ErrorCode。
//
// 系统错误（配置、网络、上游 5xx 等）通过 error 返回，此时 Result 可能为零值或仅含 Duration。
type Result struct {
	// Matched 是否核验为同一人。
	Matched bool
	// Provider 实际提供方编码。
	Provider string
	// ErrorCode 业务错误码（厂商原始码或通用 Code*）。
	ErrorCode string
	// ErrorMessage 可读错误说明（可直接用于日志；展示给用户前建议再映射）。
	ErrorMessage string
	// ProviderCode 厂商原始业务码（可选，便于对账）。
	ProviderCode string
	// RequestID 厂商请求 ID（可选）。
	RequestID string
	// Duration 本次调用耗时。
	Duration time.Duration
}

// Verifier 身份证姓名+证件号二要素核验接口。
//
// 实现应当并发安全。业务不匹配返回 (Result{Matched:false}, nil)；
// 配置缺失、网络失败、未知上游错误返回 error。
type Verifier interface {
	// Provider 返回提供方编码。
	Provider() string
	// Verify 核验姓名与身份证号是否一致。
	Verify(ctx context.Context, req Request) (Result, error)
}
