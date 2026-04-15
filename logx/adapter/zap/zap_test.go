package zap_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/logx"
	"github.com/f2xme/gox/logx/adapter/zap"
)

func TestNew_BasicLogging(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	logger := zap.New(zap.WithInfoLevel())
	logger.Info("hello", logx.NewKV("key", "value"))

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", output)
	}
	if !strings.Contains(output, "value") {
		t.Errorf("expected output to contain 'value', got %q", output)
	}
}

func TestNew_DebugLevel_Filters(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	logger := zap.New(zap.WithErrorLevel())
	logger.Info("should-not-appear")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if strings.Contains(output, "should-not-appear") {
		t.Error("info message should be filtered at error level")
	}
}

func TestNew_DisableConsole(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	logger := zap.New(zap.WithDisableConsole(), zap.WithInfoLevel())
	logger.Info("invisible")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	if n > 0 {
		t.Errorf("expected no console output, got %q", string(buf[:n]))
	}
}

func TestNew_FileOutput(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	logger := zap.New(
		zap.WithFileRotation(&zap.FileOption{
			Filename: logFile,
			MaxSize:  1,
		}),
		zap.WithDisableConsole(),
		zap.WithInfoLevel(),
	)
	logger.Info("file-test", logx.NewKV("num", 42))

	if s, ok := logger.(logx.Syncer); ok {
		s.Sync()
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "file-test") {
		t.Errorf("expected log file to contain 'file-test', got %q", string(data))
	}
}

func TestNew_AsyncBuffer(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "async.log")

	logger := zap.New(
		zap.WithFileRotation(&zap.FileOption{
			Filename: logFile,
			MaxSize:  1,
		}),
		zap.WithDisableConsole(),
		zap.WithAsyncBuffer(),
		zap.WithFlushInterval(100*time.Millisecond),
		zap.WithInfoLevel(),
	)
	logger.Info("async-test")

	// Data may not be on disk yet (buffered)
	data, _ := os.ReadFile(logFile)
	if strings.Contains(string(data), "async-test") {
		return
	}

	// Flush explicitly
	if f, ok := logger.(logx.Flusher); ok {
		f.Flush()
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "async-test") {
		t.Errorf("expected 'async-test' after flush, got %q", string(data))
	}

	if s, ok := logger.(logx.Stopper); ok {
		s.Stop()
	}
}

func TestNew_JSONFormat(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	logger := zap.New(zap.WithInfoLevel())
	logger.Info("json-test", logx.NewKV("count", 5))

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)

	var entry map[string]any
	if err := json.Unmarshal(buf[:n], &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v, output: %q", err, string(buf[:n]))
	}
	if entry["msg"] != "json-test" {
		t.Errorf("expected msg 'json-test', got %v", entry["msg"])
	}
}

func TestNew_TimeLayout(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	logger := zap.New(
		zap.WithInfoLevel(),
		zap.WithTimeLayout(time.RFC3339),
	)
	logger.Info("time-test")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "T") {
		t.Errorf("expected RFC3339 time format with 'T', got %q", output)
	}
}

func TestNewLoggers_MultiOutput(t *testing.T) {
	dir := t.TempDir()
	infoFile := filepath.Join(dir, "info.log")
	errorFile := filepath.Join(dir, "error.log")

	logger := zap.NewLoggers(
		[]zap.Option{
			zap.WithFileRotation(&zap.FileOption{Filename: infoFile}),
			zap.WithDisableConsole(),
			zap.WithInfoLevel(),
		},
		[]zap.Option{
			zap.WithFileRotation(&zap.FileOption{Filename: errorFile}),
			zap.WithDisableConsole(),
			zap.WithErrorLevel(),
		},
	)

	logger.Info("info-msg")
	logger.Error(fmt.Errorf("error-msg"))

	if s, ok := logger.(logx.Syncer); ok {
		s.Sync()
	}

	infoData, _ := os.ReadFile(infoFile)
	errorData, _ := os.ReadFile(errorFile)

	if !strings.Contains(string(infoData), "info-msg") {
		t.Errorf("info.log should contain 'info-msg', got %q", string(infoData))
	}
	if strings.Contains(string(errorData), "info-msg") {
		t.Error("error.log should not contain 'info-msg'")
	}
	if !strings.Contains(string(errorData), "error-msg") {
		t.Errorf("error.log should contain 'error-msg', got %q", string(errorData))
	}
}

func TestImplementsInterfaces(t *testing.T) {
	logger := zap.New()
	if _, ok := logger.(logx.Flusher); !ok {
		t.Error("expected Flusher interface")
	}
	if _, ok := logger.(logx.Syncer); !ok {
		t.Error("expected Syncer interface")
	}
	if _, ok := logger.(logx.Stopper); !ok {
		t.Error("expected Stopper interface")
	}
}
