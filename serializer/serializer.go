package serializer

import "io"

// Serializer 定义统一的序列化接口
type Serializer interface {
	// Marshal 将对象序列化为字节数组
	Marshal(v any) ([]byte, error)

	// Unmarshal 将字节数组反序列化为对象
	Unmarshal(data []byte, v any) error

	// Encode 将对象编码到 Writer
	Encode(w io.Writer, v any) error

	// Decode 从 Reader 解码对象
	Decode(r io.Reader, v any) error

	// ContentType 返回该序列化器的 MIME 类型
	ContentType() string
}
