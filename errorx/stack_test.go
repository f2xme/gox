package errorx

import (
	"strings"
	"testing"
)

func TestCaptureStack(t *testing.T) {
	stack := captureStack(1)
	if len(stack) == 0 {
		t.Error("expected non-empty stack")
	}

	// Check that stack frames have valid data
	if len(stack) > 0 {
		frame := stack[0]
		if frame.File == "" {
			t.Error("expected non-empty file")
		}
		if frame.Line == 0 {
			t.Error("expected non-zero line")
		}
		if frame.Function == "" {
			t.Error("expected non-empty function")
		}
	}
}

func TestErrorHasStack(t *testing.T) {
	err := New("test error")
	if len(err.Stack) == 0 {
		t.Error("expected error to have stack trace")
	}

	// Stack should contain this test function
	found := false
	for _, frame := range err.Stack {
		if strings.Contains(frame.Function, "TestErrorHasStack") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected stack to contain test function")
	}
}

func TestStackFrameString(t *testing.T) {
	frame := StackFrame{
		File:     "/path/to/file.go",
		Line:     42,
		Function: "main.TestFunc",
	}

	s := frame.String()
	if !strings.Contains(s, "file.go") {
		t.Errorf("expected string to contain 'file.go', got %q", s)
	}
	if !strings.Contains(s, "42") {
		t.Errorf("expected string to contain '42', got %q", s)
	}
	if !strings.Contains(s, "main.TestFunc") {
		t.Errorf("expected string to contain 'main.TestFunc', got %q", s)
	}
}
