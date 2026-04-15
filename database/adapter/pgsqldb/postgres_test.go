package pgsqldb

import (
	"testing"
)

// TestNew_InvalidDSN 测试无效 DSN 返回错误
func TestNew_InvalidDSN(t *testing.T) {
	_, err := New("invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

// TestMustNew_Panic 测试 MustNew 在无效 DSN 时 panic
func TestMustNew_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on invalid DSN")
		}
	}()
	MustNew("invalid-dsn")
}
