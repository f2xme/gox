package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeLogWriterClient struct {
	calls []fakeLogWriterCall
	err   error
}

type fakeLogWriterCall struct {
	index   string
	docs    []Document
	refresh bool
}

func (c *fakeLogWriterClient) CreateDoc(context.Context, string, Document, ...WriteOption) error {
	return nil
}

func (c *fakeLogWriterClient) CreateBulk(ctx context.Context, index string, docs []Document, opts ...WriteOption) error {
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("context deadline missing")
	}
	c.calls = append(c.calls, fakeLogWriterCall{
		index:   index,
		docs:    append([]Document(nil), docs...),
		refresh: ApplyWriteOptions(opts...).Refresh,
	})
	return c.err
}

func (c *fakeLogWriterClient) UpdateDoc(context.Context, string, Document, ...WriteOption) error {
	return nil
}

func (c *fakeLogWriterClient) DeleteDoc(context.Context, string, string, ...WriteOption) error {
	return nil
}

func TestNewLogWriterValidation(t *testing.T) {
	if _, err := NewLogWriter(nil, WithLogIndex("logs")); err == nil {
		t.Fatalf("NewLogWriter(nil) error = nil, want error")
	}

	client := &fakeLogWriterClient{}
	if _, err := NewLogWriter(client); err == nil {
		t.Fatalf("NewLogWriter(without index) error = nil, want error")
	}
	if _, err := NewLogWriter(client, WithLogIndex("logs"), WithLogBatchSize(10), WithLogMaxBufferedDocs(9)); err == nil {
		t.Fatal("NewLogWriter(max below batch) error = nil, want error")
	}
}

func TestLogWriterSyncFlushesJSONLogs(t *testing.T) {
	client := &fakeLogWriterClient{}
	writer, err := NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogIDFunc(func(entry map[string]any) string {
			return entry["msg"].(string)
		}),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}

	if n, err := writer.Write([]byte(`{"level":"info","msg":"hello","count":2}` + "\n")); err != nil || n == 0 {
		t.Fatalf("Write() n=%d err=%v", n, err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("CreateBulk called before Sync: %#v", client.calls)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if len(client.calls) != 1 {
		t.Fatalf("CreateBulk calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.index != "app-logs" {
		t.Fatalf("index = %q, want app-logs", call.index)
	}
	if len(call.docs) != 1 || call.docs[0].ID() != "hello" {
		t.Fatalf("docs = %#v, want one doc with ID hello", call.docs)
	}

	var body map[string]any
	data, err := json.Marshal(call.docs[0])
	if err != nil {
		t.Fatalf("marshal doc: %v", err)
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal doc: %v", err)
	}
	if body["msg"] != "hello" || body["level"] != "info" {
		t.Fatalf("doc body = %#v", body)
	}
}

func TestLogWriterBatchSizeFlushesFromWrite(t *testing.T) {
	client := &fakeLogWriterClient{}
	writer, err := NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogBatchSize(2),
		WithLogRefresh(true),
		WithLogFlushTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}

	payload := []byte(`{"msg":"one"}` + "\n" + `{"msg":"two"}` + "\n")
	if n, err := writer.Write(payload); err != nil || n != len(payload) {
		t.Fatalf("Write() n=%d err=%v, want n=%d nil", n, err, len(payload))
	}

	if len(client.calls) != 1 {
		t.Fatalf("CreateBulk calls = %d, want 1", len(client.calls))
	}
	if !client.calls[0].refresh {
		t.Fatalf("refresh = false, want true")
	}
	if len(client.calls[0].docs) != 2 {
		t.Fatalf("docs = %d, want 2", len(client.calls[0].docs))
	}
}

func TestLogWriterFlushErrorIsReportedAndRetained(t *testing.T) {
	client := &fakeLogWriterClient{err: errors.New("bulk failed")}
	var handled error
	writer, err := NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogErrorHandler(func(err error) {
			handled = err
		}),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}
	if _, err := writer.Write([]byte(`{"msg":"retry-me"}` + "\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := writer.Flush(); err == nil {
		t.Fatalf("Flush() error = nil, want bulk error")
	}
	if handled == nil {
		t.Fatalf("error handler was not called")
	}

	client.err = nil
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() retry error = %v", err)
	}
	if len(client.calls) != 2 {
		t.Fatalf("CreateBulk calls = %d, want 2", len(client.calls))
	}
}

func TestLogWriterCloseCanRetryAfterFlushError(t *testing.T) {
	client := &fakeLogWriterClient{err: errors.New("bulk failed")}
	writer, err := NewLogWriter(client, WithLogIndex("app-logs"))
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}
	if _, err := writer.Write([]byte(`{"msg":"retry-close"}` + "\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := writer.Close(); err == nil {
		t.Fatal("first Close() error = nil, want bulk error")
	}
	client.err = nil
	if err := writer.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if len(client.calls) != 2 {
		t.Fatalf("CreateBulk calls = %d, want 2", len(client.calls))
	}
	if _, err := writer.Write([]byte(`{"msg":"closed"}` + "\n")); err == nil {
		t.Fatal("Write() after successful Close error = nil, want error")
	}
}

func TestLogWriterErrorHandlerRunsWithoutWriterLock(t *testing.T) {
	client := &fakeLogWriterClient{err: errors.New("bulk failed")}
	var writer *LogWriter
	var handlerErr error
	var handled bool
	var err error
	writer, err = NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogErrorHandler(func(error) {
			handled = true
			client.err = nil
			_, handlerErr = writer.Write([]byte(`{"msg":"handler"}` + "\n"))
		}),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}
	if _, err := writer.Write([]byte(`{"msg":"original"}` + "\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := writer.Flush(); err == nil {
		t.Fatal("Flush() error = nil, want bulk error")
	}
	if !handled || handlerErr != nil {
		t.Fatalf("handler ran = %v, handler error = %v", handled, handlerErr)
	}
}

func TestLogWriterErrorHandlerDoesNotRecurse(t *testing.T) {
	client := &fakeLogWriterClient{err: errors.New("bulk failed")}
	var writer *LogWriter
	handled := 0
	var err error
	writer, err = NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogBatchSize(1),
		WithLogMaxBufferedDocs(2),
		WithLogErrorHandler(func(error) {
			handled++
			_, _ = writer.Write([]byte(`{"msg":"handler"}` + "\n"))
		}),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}

	if _, err := writer.Write([]byte(`{"msg":"original"}` + "\n")); err == nil {
		t.Fatal("Write() error = nil, want bulk error")
	}
	if handled != 1 {
		t.Fatalf("handler calls = %d, want 1", handled)
	}
}

func TestLogWriterBufferIsBounded(t *testing.T) {
	client := &fakeLogWriterClient{err: errors.New("bulk failed")}
	writer, err := NewLogWriter(client,
		WithLogIndex("app-logs"),
		WithLogBatchSize(2),
		WithLogMaxBufferedDocs(3),
	)
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}

	if _, err := writer.Write([]byte(`{"msg":"one"}` + "\n" + `{"msg":"two"}` + "\n")); err == nil {
		t.Fatal("first Write() error = nil, want bulk error")
	}
	if _, err := writer.Write([]byte(`{"msg":"three"}` + "\n")); err == nil {
		t.Fatal("second Write() error = nil, want bulk error")
	}
	if n, err := writer.Write([]byte(`{"msg":"overflow"}` + "\n")); err == nil || n != 0 {
		t.Fatalf("overflow Write() n=%d err=%v, want 0 and error", n, err)
	}
	if got := len(writer.docs); got != 3 {
		t.Fatalf("buffered docs = %d, want 3", got)
	}
	client.err = nil
	if n, err := writer.Write([]byte(`{"msg":"after-recovery"}` + "\n")); err != nil || n == 0 {
		t.Fatalf("recovery Write() n=%d err=%v", n, err)
	}
	if got := len(writer.docs); got != 1 {
		t.Fatalf("buffered docs after recovery = %d, want 1", got)
	}
}

func TestLogWriterRejectsInvalidJSON(t *testing.T) {
	client := &fakeLogWriterClient{}
	writer, err := NewLogWriter(client, WithLogIndex("app-logs"))
	if err != nil {
		t.Fatalf("NewLogWriter() error = %v", err)
	}

	if _, err := writer.Write([]byte(`not-json`)); err == nil {
		t.Fatalf("Write(invalid) error = nil, want error")
	}
}
