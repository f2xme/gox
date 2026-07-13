package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultLogWriterBatchSize    = 100
	defaultLogWriterFlushTimeout = 5 * time.Second
)

var logWriterSeq uint64

// LogIDFunc 定义日志文档 ID 生成函数。
type LogIDFunc func(entry map[string]any) string

// LogWriterOptions 定义日志写入器选项。
type LogWriterOptions struct {
	// Index 目标日志索引。
	Index string
	// BatchSize 触发批量写入的日志条数。
	BatchSize int
	// MaxBufferedDocs 内存中最多保留的日志条数，默认等于 BatchSize。
	MaxBufferedDocs int
	// FlushTimeout 单次批量写入超时时间。
	FlushTimeout time.Duration
	// Refresh 是否在写入后立即刷新索引。
	Refresh bool
	// ErrorHandler 处理异步 flush 或 close 时遇到的错误。
	ErrorHandler func(error)
	// IDFunc 生成日志文档 ID。
	IDFunc LogIDFunc
}

// LogWriterOption 定义日志写入器选项函数。
type LogWriterOption func(*LogWriterOptions)

// LogWriter 将 JSON 日志行批量写入 Elasticsearch。
//
// LogWriter 实现 io.Writer，并提供 Sync、Flush 和 Close 方法，可直接传给
// zapcore.AddSync 或 logx/adapter/zap.WithWriter。
type LogWriter struct {
	client Writer

	index           string
	batchSize       int
	maxBufferedDocs int
	flushTimeout    time.Duration
	refresh         bool
	errorHandler    func(error)
	idFunc          LogIDFunc

	mu            sync.Mutex
	docs          []Document
	closed        bool
	handlingError atomic.Bool
}

// NewLogWriter 创建写入 Elasticsearch 的日志 writer。
func NewLogWriter(client Writer, opts ...LogWriterOption) (*LogWriter, error) {
	if client == nil {
		return nil, fmt.Errorf("elastic: log writer client is nil")
	}

	o := defaultLogWriterOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if strings.TrimSpace(o.Index) == "" {
		return nil, fmt.Errorf("elastic: log writer index is required")
	}
	if o.MaxBufferedDocs == 0 {
		o.MaxBufferedDocs = o.BatchSize
	}
	if o.MaxBufferedDocs < o.BatchSize {
		return nil, fmt.Errorf("elastic: max buffered docs must be at least batch size")
	}

	return &LogWriter{
		client:          client,
		index:           o.Index,
		batchSize:       o.BatchSize,
		maxBufferedDocs: o.MaxBufferedDocs,
		flushTimeout:    o.FlushTimeout,
		refresh:         o.Refresh,
		errorHandler:    o.ErrorHandler,
		idFunc:          o.IDFunc,
		docs:            make([]Document, 0, o.BatchSize),
	}, nil
}

func defaultLogWriterOptions() LogWriterOptions {
	return LogWriterOptions{
		BatchSize:    defaultLogWriterBatchSize,
		FlushTimeout: defaultLogWriterFlushTimeout,
		IDFunc:       defaultLogID,
	}
}

// WithLogIndex 设置日志索引。
func WithLogIndex(index string) LogWriterOption {
	return func(o *LogWriterOptions) {
		o.Index = index
	}
}

// WithLogBatchSize 设置触发批量写入的日志条数。
func WithLogBatchSize(size int) LogWriterOption {
	return func(o *LogWriterOptions) {
		if size > 0 {
			o.BatchSize = size
		}
	}
}

// WithLogMaxBufferedDocs 设置内存中最多保留的日志条数。
//
// 达到上限后，Write 返回错误且不接收新日志，避免 Elasticsearch 故障期间
// 缓冲无限增长。该值必须不小于 BatchSize。
func WithLogMaxBufferedDocs(size int) LogWriterOption {
	return func(o *LogWriterOptions) {
		if size > 0 {
			o.MaxBufferedDocs = size
		}
	}
}

// WithLogFlushTimeout 设置单次批量写入超时时间。
func WithLogFlushTimeout(timeout time.Duration) LogWriterOption {
	return func(o *LogWriterOptions) {
		if timeout > 0 {
			o.FlushTimeout = timeout
		}
	}
}

// WithLogRefresh 设置写入后是否立即刷新索引。
func WithLogRefresh(refresh bool) LogWriterOption {
	return func(o *LogWriterOptions) {
		o.Refresh = refresh
	}
}

// WithLogErrorHandler 设置日志写入错误处理函数。
func WithLogErrorHandler(handler func(error)) LogWriterOption {
	return func(o *LogWriterOptions) {
		o.ErrorHandler = handler
	}
}

// WithLogIDFunc 设置日志文档 ID 生成函数。
func WithLogIDFunc(fn LogIDFunc) LogWriterOption {
	return func(o *LogWriterOptions) {
		if fn != nil {
			o.IDFunc = fn
		}
	}
}

// Write 写入一批 JSON 日志行。
func (w *LogWriter) Write(p []byte) (int, error) {
	docs, err := w.parseDocs(p)
	if err != nil {
		return 0, err
	}
	if len(docs) == 0 {
		return len(p), nil
	}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return 0, fmt.Errorf("elastic: log writer is closed")
	}
	if len(docs) > w.maxBufferedDocs {
		w.mu.Unlock()
		return 0, fmt.Errorf("elastic: log buffer is full: max %d documents", w.maxBufferedDocs)
	}
	if len(docs) > w.maxBufferedDocs-len(w.docs) {
		err = w.flushLocked()
		if err != nil {
			w.mu.Unlock()
			w.handleError(err)
			return 0, err
		}
	}
	w.docs = append(w.docs, docs...)
	if len(w.docs) < w.batchSize {
		w.mu.Unlock()
		return len(p), nil
	}
	err = w.flushLocked()
	w.mu.Unlock()
	if err != nil {
		w.handleError(err)
		return len(p), err
	}
	return len(p), nil
}

// Sync 刷新已缓存的日志。
func (w *LogWriter) Sync() error {
	return w.Flush()
}

// Flush 刷新已缓存的日志。
func (w *LogWriter) Flush() error {
	w.mu.Lock()
	err := w.flushLocked()
	w.mu.Unlock()
	if err != nil {
		w.handleError(err)
	}
	return err
}

// Close 刷新已缓存的日志并关闭 writer。
func (w *LogWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	err := w.flushLocked()
	if err == nil {
		w.closed = true
	}
	w.mu.Unlock()
	if err != nil {
		w.handleError(err)
	}
	return err
}

func (w *LogWriter) parseDocs(p []byte) ([]Document, error) {
	lines := bytes.Split(p, []byte{'\n'})
	docs := make([]Document, 0, len(lines))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("elastic: decode log json: %w", err)
		}
		if entry == nil {
			return nil, fmt.Errorf("elastic: log json must be an object")
		}
		docs = append(docs, logDocument{
			id:     w.idFunc(entry),
			fields: entry,
		})
	}
	return docs, nil
}

func (w *LogWriter) flushLocked() error {
	if len(w.docs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), w.flushTimeout)
	defer cancel()

	opts := []WriteOption(nil)
	if w.refresh {
		opts = append(opts, WithRefresh(true))
	}

	docs := append([]Document(nil), w.docs...)
	if err := w.client.CreateBulk(ctx, w.index, docs, opts...); err != nil {
		return fmt.Errorf("elastic: flush log writer: %w", err)
	}
	w.docs = w.docs[:0]
	return nil
}

func (w *LogWriter) handleError(err error) {
	if w.errorHandler == nil || !w.handlingError.CompareAndSwap(false, true) {
		return
	}
	defer w.handlingError.Store(false)
	w.errorHandler(err)
}

type logDocument struct {
	id     string
	fields map[string]any
}

func (d logDocument) ID() string {
	return d.id
}

func (d logDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.fields)
}

func defaultLogID(map[string]any) string {
	seq := atomic.AddUint64(&logWriterSeq, 1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), seq)
}
