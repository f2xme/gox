package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/idverify"
)

// Verifier 是并发安全的内存二要素核验实现。
type Verifier struct {
	mu      sync.Mutex
	options Options
	calls   []idverify.Request
}

var _ idverify.Verifier = (*Verifier)(nil)

// Provider 返回 mock。
func (v *Verifier) Provider() string { return idverify.ProviderMock }

// Verify 按配置返回匹配结果。
func (v *Verifier) Verify(ctx context.Context, req idverify.Request) (idverify.Result, error) {
	start := time.Now()
	if ctx == nil {
		return idverify.Result{Duration: time.Since(start)}, fmt.Errorf("%w: context is nil", idverify.ErrInvalidArgument)
	}
	if err := ctx.Err(); err != nil {
		return idverify.Result{Duration: time.Since(start)}, err
	}

	req = req.Normalize()
	if req.Name == "" || req.IDNumber == "" {
		return idverify.Result{Provider: idverify.ProviderMock, Duration: time.Since(start)},
			fmt.Errorf("%w: name and id number are required", idverify.ErrInvalidArgument)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls = append(v.calls, req)

	if v.options.VerifyError != nil {
		return idverify.Result{Provider: idverify.ProviderMock, Duration: time.Since(start)}, v.options.VerifyError
	}

	for _, name := range v.options.MismatchNames {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if strings.TrimSpace(name) == req.Name {
			return idverify.Result{
				Matched:      false,
				Provider:     idverify.ProviderMock,
				ErrorCode:    idverify.CodeNameMismatch,
				ErrorMessage: "name and id number mismatch",
				Duration:     time.Since(start),
			}, nil
		}
	}
	for _, name := range v.options.InvalidIDNames {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if strings.TrimSpace(name) == req.Name {
			return idverify.Result{
				Matched:      false,
				Provider:     idverify.ProviderMock,
				ErrorCode:    idverify.CodeIDInvalid,
				ErrorMessage: "id number invalid or not found",
				Duration:     time.Since(start),
			}, nil
		}
	}

	return idverify.Result{
		Matched:  true,
		Provider: idverify.ProviderMock,
		Duration: time.Since(start),
	}, nil
}

// Calls 返回已调用请求副本。
func (v *Verifier) Calls() []idverify.Request {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]idverify.Request, len(v.calls))
	copy(out, v.calls)
	return out
}

// Reset 清空调用记录。
func (v *Verifier) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls = v.calls[:0]
}
