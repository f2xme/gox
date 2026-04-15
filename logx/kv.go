package logx

// KV 实现 Meta 接口的键值对结构
type KV struct {
	key   string
	value any
}

// NewKV 创建一个新的键值对 Meta
//
// 示例：
//
//	meta := logx.NewKV("user_id", "123")
//	logger.Info("user login", meta)
func NewKV(key string, value any) *KV {
	return &KV{key: key, value: value}
}

func (m *KV) Key() string { return m.key }

func (m *KV) Value() any { return m.value }
