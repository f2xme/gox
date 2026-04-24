package pgsqldb

import (
	"testing"
)

// TestNew_InvalidDSN 测试无效 DSN 返回错误
func TestNew_InvalidDSN(t *testing.T) {
	_, err := New(WithDSN("invalid-dsn"))
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}
