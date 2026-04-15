module github.com/f2xme/gox/logx/adapter/zap

go 1.25.7

replace github.com/f2xme/gox => ../../../

require (
	github.com/f2xme/gox v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require go.uber.org/multierr v1.10.0 // indirect
