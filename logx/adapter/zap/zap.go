package zap

import (
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
	mu              sync.Mutex
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
	var lastErr error
	for _, bw := range l.bufferedWriters {
		if err := bw.Sync(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *zapLogger) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var lastErr error
	for _, bw := range l.bufferedWriters {
		if err := bw.Stop(); err != nil {
			lastErr = err
		}
	}
	return lastErr
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

	if cfg.File != nil {
		var ws zapcore.WriteSyncer
		if cfg.Buffer != nil {
			bws := &zapcore.BufferedWriteSyncer{
				WS:            zapcore.AddSync(cfg.File),
				Size:          cfg.Buffer.Size,
				FlushInterval: cfg.Buffer.Interval,
			}
			result.bufferedWriters = append(result.bufferedWriters, bws)
			ws = bws
		} else {
			ws = zapcore.AddSync(cfg.File)
		}
		cores = append(cores, zapcore.NewCore(jsonEncoder, ws, levelEnabler))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	result.core = zapcore.NewTee(cores...)
	return result
}
