package logx

import "testing"

func TestNewKV(t *testing.T) {
	kv := NewKV("user_id", 123)
	if kv.Key() != "user_id" {
		t.Errorf("expected key 'user_id', got %q", kv.Key())
	}
	if kv.Value() != 123 {
		t.Errorf("expected value 123, got %v", kv.Value())
	}
}

func TestKV_ImplementsMeta(t *testing.T) {
	var m Meta = NewKV("key", "value")
	if m.Key() != "key" {
		t.Error("KV should implement Meta")
	}
}
