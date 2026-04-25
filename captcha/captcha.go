package captcha

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Service 定义验证码服务接口。
type Service interface {
	// Generate 生成验证码。
	Generate(ctx context.Context) (Challenge, error)

	// Verify 验证验证码答案。
	// 验证成功返回 true，失败返回 false。
	// 验证成功后会自动删除验证码，防止重复使用。
	Verify(ctx context.Context, id string, answer string) (bool, error)

	// Delete 删除验证码。
	// 通常在验证成功后调用，防止重复使用。
	Delete(ctx context.Context, id string) error

	// Regenerate 重新生成验证码内容（保持相同 ID）。
	// 用于"看不清，换一张"的场景。
	Regenerate(ctx context.Context, id string) (Challenge, error)
}

// Captcha 是 Service 的兼容别名。
type Captcha = Service

type service struct {
	store     Store
	generator Generator
	opts      Options
}

// Generate 生成验证码。
func (s *service) Generate(ctx context.Context) (Challenge, error) {
	// 生成随机 ID
	id, err := generateID(s.opts.IDLength)
	if err != nil {
		return Challenge{}, err
	}

	// 生成验证码
	data, err := s.generator.Generate(ctx)
	if err != nil {
		return Challenge{}, fmt.Errorf("%w: %w", ErrGenerateFailed, err)
	}

	// 存储答案
	if err := s.store.Set(ctx, id, data.Answer, s.opts.TTL); err != nil {
		return Challenge{}, err
	}

	return Challenge{
		ID:   id,
		Data: data.Data,
		Type: s.generator.Type(),
	}, nil
}

// Verify 验证验证码答案。
func (s *service) Verify(ctx context.Context, id string, answer string) (bool, error) {
	if id == "" || answer == "" {
		return false, nil
	}

	stored, err := s.getAnswer(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	// 比较答案（忽略大小写和空格）
	ok := compareAnswer(stored, answer)
	if s.shouldDelete(ok) {
		if err := s.store.Delete(ctx, id); err != nil {
			return false, err
		}
	}

	return ok, nil
}

// Delete 删除验证码。
func (s *service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidID
	}
	return s.store.Delete(ctx, id)
}

// Regenerate 重新生成验证码内容（保持相同 ID）。
func (s *service) Regenerate(ctx context.Context, id string) (Challenge, error) {
	if id == "" {
		return Challenge{}, ErrInvalidID
	}
	if _, err := s.store.Get(ctx, id); err != nil {
		return Challenge{}, err
	}

	// 生成新验证码
	data, err := s.generator.Generate(ctx)
	if err != nil {
		return Challenge{}, fmt.Errorf("%w: %w", ErrGenerateFailed, err)
	}

	// 更新答案
	if err := s.store.Set(ctx, id, data.Answer, s.opts.TTL); err != nil {
		return Challenge{}, err
	}

	return Challenge{
		ID:   id,
		Data: data.Data,
		Type: s.generator.Type(),
	}, nil
}

// getAnswer 根据消费策略获取答案。
func (s *service) getAnswer(ctx context.Context, id string) (string, error) {
	if s.opts.ConsumeMode == ConsumeAlways {
		if taker, ok := s.store.(Taker); ok {
			return taker.Take(ctx, id)
		}
	}
	return s.store.Get(ctx, id)
}

// shouldDelete 判断当前验证结果是否需要删除验证码。
func (s *service) shouldDelete(ok bool) bool {
	switch s.opts.ConsumeMode {
	case ConsumeAlways:
		if _, supportsTake := s.store.(Taker); supportsTake {
			return false
		}
		return true
	case ConsumeOnSuccess:
		return ok
	default:
		return false
	}
}

// generateID 生成随机 ID。
func generateID(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidID
	}

	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// compareAnswer 比较答案（忽略大小写和空格）。
func compareAnswer(stored, input string) bool {
	stored = strings.TrimSpace(strings.ToLower(stored))
	input = strings.TrimSpace(strings.ToLower(input))
	return stored == input
}
