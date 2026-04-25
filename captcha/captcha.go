package captcha

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/f2xme/gox/captcha/generator"
)

// Captcha 定义验证码服务接口。
type Captcha interface {
	// Generate 生成验证码，返回 ID 和数据。
	Generate(ctx context.Context) (id string, data string, err error)

	// Verify 验证验证码答案。
	// 验证成功返回 true，失败返回 false。
	// 不会自动删除验证码，需要手动调用 Delete。
	Verify(ctx context.Context, id string, answer string) (bool, error)

	// Delete 删除验证码。
	// 通常在验证成功后调用，防止重复使用。
	Delete(ctx context.Context, id string) error

	// Regenerate 重新生成验证码内容（保持相同 ID）。
	// 用于"看不清，换一张"的场景。
	Regenerate(ctx context.Context, id string) (data string, err error)
}

type captcha struct {
	store     Store
	generator generator.Generator
	opts      Options
}

// Generate 生成验证码。
func (c *captcha) Generate(ctx context.Context) (string, string, error) {
	// 生成随机 ID
	id, err := generateID(c.opts.IDLength)
	if err != nil {
		return "", "", err
	}

	// 生成验证码
	data, answer, err := c.generator.Generate()
	if err != nil {
		return "", "", ErrGenerateFailed
	}

	// 存储答案
	if err := c.store.Set(ctx, id, answer, c.opts.TTL); err != nil {
		return "", "", err
	}

	return id, data, nil
}

// Verify 验证验证码答案。
func (c *captcha) Verify(ctx context.Context, id string, answer string) (bool, error) {
	if id == "" || answer == "" {
		return false, nil
	}

	// 获取存储的答案
	stored, err := c.store.Get(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		return false, err
	}

	// 比较答案（忽略大小写和空格）
	return compareAnswer(stored, answer), nil
}

// Delete 删除验证码。
func (c *captcha) Delete(ctx context.Context, id string) error {
	return c.store.Delete(ctx, id)
}

// Regenerate 重新生成验证码内容（保持相同 ID）。
func (c *captcha) Regenerate(ctx context.Context, id string) (string, error) {
	// 生成新验证码
	data, answer, err := c.generator.Generate()
	if err != nil {
		return "", ErrGenerateFailed
	}

	// 更新答案
	if err := c.store.Set(ctx, id, answer, c.opts.TTL); err != nil {
		return "", err
	}

	return data, nil
}

// generateID 生成随机 ID。
func generateID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// compareAnswer 比较答案（忽略大小写和空格）。
func compareAnswer(stored, input string) bool {
	stored = strings.TrimSpace(strings.ToLower(stored))
	input = strings.TrimSpace(strings.ToLower(input))
	return stored == input
}
