package elasticsearch

// WriteOptions 定义文档写入选项。
type WriteOptions struct {
	// Refresh 是否在写入后立即刷新索引。
	Refresh bool
}

// WriteOption 定义文档写入选项函数。
type WriteOption func(*WriteOptions)

func defaultWriteOptions() WriteOptions {
	return WriteOptions{}
}

// ApplyWriteOptions 应用文档写入选项。
func ApplyWriteOptions(opts ...WriteOption) WriteOptions {
	o := defaultWriteOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// WithRefresh 设置是否在写入后立即刷新索引。
//
// 示例：
//
//	client.CreateDoc(ctx, "users", user, WithRefresh(true))
func WithRefresh(refresh bool) WriteOption {
	return func(o *WriteOptions) {
		o.Refresh = refresh
	}
}
