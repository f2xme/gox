package zap

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/f2xme/gox/logx"
	gozap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger          *gozap.Logger
	bufferedWriters []*zapcore.BufferedWriteSyncer
	outputFlushers  []func() error
	outputClosers   []io.Closer
	mu              sync.Mutex
	stopOnce        sync.Once
	stopErr         error
}

var (
	_ logx.Logger  = (*zapLogger)(nil)
	_ logx.Flusher = (*zapLogger)(nil)
	_ logx.Syncer  = (*zapLogger)(nil)
	_ logx.Stopper = (*zapLogger)(nil)
)

func (l *zapLogger) Info(msg string, metas ...logx.Meta) {
	l.logger.Info(msg, toFields(metas)...)
}

func (l *zapLogger) Warn(msg string, metas ...logx.Meta) {
	l.logger.Warn(msg, toFields(metas)...)
}

func (l *zapLogger) Error(err error, metas ...logx.Meta) {
	l.logger.Error(err.Error(), toFields(metas)...)
}

func (l *zapLogger) Fatal(err error, metas ...logx.Meta) {
	l.logger.Fatal(err.Error(), toFields(metas)...)
}

func (l *zapLogger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	errs := make([]error, 0, len(l.bufferedWriters)+len(l.outputFlushers))
	for _, bw := range l.bufferedWriters {
		if err := bw.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	for _, flush := range l.outputFlushers {
		if err := flush(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (l *zapLogger) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logger.Sync()
}

func (l *zapLogger) Stop() error {
	l.stopOnce.Do(func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		errs := make([]error, 0, len(l.bufferedWriters)+len(l.outputFlushers)+len(l.outputClosers))
		for _, bw := range l.bufferedWriters {
			if err := bw.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
		for _, flush := range l.outputFlushers {
			if err := flush(); err != nil {
				errs = append(errs, err)
			}
		}
		for _, closer := range l.outputClosers {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		l.stopErr = errors.Join(errs...)
	})
	return l.stopErr
}

func toFields(metas []logx.Meta) []gozap.Field {
	if len(metas) == 0 {
		return nil
	}
	fields := make([]gozap.Field, 0, len(metas))
	for _, m := range metas {
		fields = append(fields, gozap.Any(m.Key(), m.Value()))
	}
	return fields
}

type coreResult struct {
	core            zapcore.Core
	bufferedWriters []*zapcore.BufferedWriteSyncer
	outputFlushers  []func() error
	outputClosers   []io.Closer
}

func buildCore(cfg *Options) *coreResult {
	result := &coreResult{}

	encoderCfg := gozap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(cfg.TimeLayout))
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	levelEnabler := gozap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= cfg.Level
	})

	var cores []zapcore.Core

	if !cfg.DisableConsole {
		cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), levelEnabler))
	}

	for _, writer := range outputWriters(cfg) {
		var ws zapcore.WriteSyncer
		if cfg.Buffer != nil {
			bws := &zapcore.BufferedWriteSyncer{
				WS:            zapcore.AddSync(writer),
				Size:          cfg.Buffer.Size,
				FlushInterval: cfg.Buffer.Interval,
			}
			result.bufferedWriters = append(result.bufferedWriters, bws)
			ws = bws
		} else {
			ws = zapcore.AddSync(writer)
		}
		appendOutputLifecycle(result, writer, cfg.Buffer != nil)
		cores = append(cores, zapcore.NewCore(jsonEncoder, ws, levelEnabler))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	result.core = zapcore.NewTee(cores...)
	return result
}

func outputWriters(cfg *Options) []io.Writer {
	writers := make([]io.Writer, 0, len(cfg.Writers)+1)
	if cfg.File != nil {
		writers = append(writers, cfg.File)
	}
	writers = append(writers, cfg.Writers...)
	return writers
}

func appendOutputLifecycle(result *coreResult, writer io.Writer, buffered bool) {
	f, hasFlush := writer.(interface{ Flush() error })
	s, hasSync := writer.(interface{ Sync() error })
	if buffered {
		if hasFlush && !hasSync {
			result.outputFlushers = append(result.outputFlushers, f.Flush)
		}
	} else if hasFlush {
		result.outputFlushers = append(result.outputFlushers, f.Flush)
	} else if hasSync {
		result.outputFlushers = append(result.outputFlushers, s.Sync)
	}
	if c, ok := writer.(io.Closer); ok {
		result.outputClosers = append(result.outputClosers, c)
	}
}
