package zap

// FileOption 定义日志文件轮转配置
type FileOption struct {
	// Filename 日志文件路径
	Filename string
	// MaxSize 单个日志文件最大大小（MB），默认 10MB
	MaxSize int
	// MaxBackups 保留的旧日志文件最大数量，默认 10
	MaxBackups int
	// MaxAge 保留旧日志文件的最大天数，默认 10 天
	MaxAge int
	// LocalTime 是否使用本地时间作为轮转时间戳，默认 false（使用 UTC）
	LocalTime bool
	// Compress 是否压缩轮转后的日志文件，默认 false
	Compress bool
}
