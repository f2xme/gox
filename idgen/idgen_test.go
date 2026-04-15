package idgen

import "testing"

func TestGeneratorFunc(t *testing.T) {
	called := false
	gen := GeneratorFunc(func() string {
		called = true
		return "test-id"
	})

	id := gen.Generate()
	if id != "test-id" {
		t.Errorf("expected 'test-id', got %q", id)
	}
	if !called {
		t.Error("generator function was not called")
	}
}
