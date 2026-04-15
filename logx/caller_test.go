package logx

import "testing"

func TestTrimCallerPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/Users/xxx/project/app/service/user.go", "app/service/user.go"},
		{"/home/ci/build/library/log/log.go", "library/log/log.go"},
		{"/go/src/internal/handler.go", "internal/handler.go"},
		{"/go/src/cmd/main.go", "cmd/main.go"},
		{"/some/random/path.go", "/some/random/path.go"},
		{"relative/path.go", "relative/path.go"},
	}
	for _, tt := range tests {
		if got := trimCallerPath(tt.input); got != tt.want {
			t.Errorf("trimCallerPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
