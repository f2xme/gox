package errorx

import (
	"testing"
)

func BenchmarkBizErrorWithMeta(b *testing.B) {
	base := NewBizWithMessage("TEST", "message: {value}")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.WithMeta("value", i).Error()
	}
}

func BenchmarkBizErrorWithoutMeta(b *testing.B) {
	err := NewBizWithMessage("TEST", "message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkBizErrorWithMetaNoPlaceholder(b *testing.B) {
	base := NewBizWithMessage("TEST", "message without placeholder")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.WithMeta("value", i).Error()
	}
}

func BenchmarkBizErrorMultipleMeta(b *testing.B) {
	base := NewBizWithMessage("TEST", "user {user_id} action {action} at {time}")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.
			WithMeta("user_id", 12345).
			WithMeta("action", "login").
			WithMeta("time", "2024-01-01").
			Error()
	}
}

func BenchmarkGetBizMeta(b *testing.B) {
	err := NewBizWithMessage("TEST", "message").
		WithMeta("key1", "value1").
		WithMeta("key2", 123)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBizMeta(err)
	}
}

func BenchmarkBizErrorLocalize(b *testing.B) {
	resetRegistryForTest()
	Register("USER_NOT_FOUND", "zh", "用户 {user_id} 不存在")
	err := NewBiz("USER_NOT_FOUND").WithMeta("user_id", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Localize("zh")
	}
}

func BenchmarkBizErrorWithMetaMap(b *testing.B) {
	base := NewBizWithMessage("TEST", "user {user_id} action {action} at {time}")
	meta := map[string]any{
		"user_id": 12345,
		"action":  "login",
		"time":    "2024-01-01",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.WithMetaMap(meta).Error()
	}
}

func BenchmarkBizErrorWithMetaMapVsChain(b *testing.B) {
	base := NewBizWithMessage("TEST", "user {user_id} action {action} at {time}")
	meta := map[string]any{
		"user_id": 12345,
		"action":  "login",
		"time":    "2024-01-01",
	}

	b.Run("WithMetaMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = base.WithMetaMap(meta).Error()
		}
	})

	b.Run("ChainedWithMeta", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = base.
				WithMeta("user_id", 12345).
				WithMeta("action", "login").
				WithMeta("time", "2024-01-01").
				Error()
		}
	})
}
