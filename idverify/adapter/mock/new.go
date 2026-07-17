package mock

// New 创建 mock 核验器。
func New(opts ...Option) (*Verifier, error) {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return &Verifier{options: o}, nil
}

// MustNew 创建 mock 核验器，失败时 panic。
func MustNew(opts ...Option) *Verifier {
	v, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return v
}
