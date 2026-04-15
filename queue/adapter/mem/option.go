package memadapter

// Options holds the configuration for memory queue.
type Options struct {
	// BufferSize is the channel buffer size for each topic subscription.
	// Default is 64.
	BufferSize int
}

// defaultOptions returns default configuration.
func defaultOptions() Options {
	return Options{
		BufferSize: 64,
	}
}

// Option is a function that modifies Options.
type Option func(*Options)

// WithBufferSize sets the channel buffer size for each topic subscription.
// A value <= 0 defaults to 64.
//
// Example:
//
//	mem.New(mem.WithBufferSize(128))
func WithBufferSize(size int) Option {
	return func(o *Options) {
		if size <= 0 {
			size = 64
		}
		o.BufferSize = size
	}
}
