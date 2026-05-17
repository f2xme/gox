package logx

import "sync"

const defaultAsyncBufferSize = 1024

type asyncEntry struct {
	level string
	msg   string
	err   error
	metas []Meta
	done  chan struct{}
}

type asyncLogger struct {
	logger Logger
	queue  chan asyncEntry

	mu      sync.Mutex
	stopped bool
	wg      sync.WaitGroup
	stopErr error
	once    sync.Once
}

func newAsyncLogger(logger Logger, bufferSize int) *asyncLogger {
	if bufferSize <= 0 {
		bufferSize = defaultAsyncBufferSize
	}
	l := &asyncLogger{
		logger: logger,
		queue:  make(chan asyncEntry, bufferSize),
	}
	l.wg.Add(1)
	go l.run()
	return l
}

func (l *asyncLogger) run() {
	defer l.wg.Done()
	for entry := range l.queue {
		l.write(entry)
	}
}

func (l *asyncLogger) enqueue(entry asyncEntry) {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		l.write(entry)
		return
	}
	l.queue <- entry
	l.mu.Unlock()
}

func (l *asyncLogger) wait() {
	done := make(chan struct{})
	l.enqueue(asyncEntry{done: done})
	<-done
}

func (l *asyncLogger) write(entry asyncEntry) {
	if entry.done != nil {
		close(entry.done)
		return
	}
	switch entry.level {
	case "info":
		l.logger.Info(entry.msg, entry.metas...)
	case "warn":
		l.logger.Warn(entry.msg, entry.metas...)
	case "error":
		l.logger.Error(entry.err, entry.metas...)
	}
}

func (l *asyncLogger) Info(msg string, metas ...Meta) {
	l.enqueue(asyncEntry{level: "info", msg: msg, metas: cloneMetas(metas)})
}

func (l *asyncLogger) Warn(msg string, metas ...Meta) {
	l.enqueue(asyncEntry{level: "warn", msg: msg, metas: cloneMetas(metas)})
}

func (l *asyncLogger) Error(err error, metas ...Meta) {
	if err == nil {
		return
	}
	l.enqueue(asyncEntry{level: "error", err: err, metas: cloneMetas(metas)})
}

func (l *asyncLogger) Fatal(err error, metas ...Meta) {
	if err == nil {
		return
	}
	l.wait()
	l.logger.Fatal(err, cloneMetas(metas)...)
}

func (l *asyncLogger) Flush() error {
	l.wait()
	if f, ok := l.logger.(Flusher); ok {
		return f.Flush()
	}
	return nil
}

func (l *asyncLogger) Sync() error {
	l.wait()
	if s, ok := l.logger.(Syncer); ok {
		return s.Sync()
	}
	return nil
}

func (l *asyncLogger) Stop() error {
	l.once.Do(func() {
		l.mu.Lock()
		l.stopped = true
		close(l.queue)
		l.mu.Unlock()

		l.wg.Wait()
		if s, ok := l.logger.(Stopper); ok {
			l.stopErr = s.Stop()
		}
	})
	return l.stopErr
}

func cloneMetas(metas []Meta) []Meta {
	if len(metas) == 0 {
		return nil
	}
	cloned := make([]Meta, len(metas))
	copy(cloned, metas)
	return cloned
}
