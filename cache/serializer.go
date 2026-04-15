package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// Serializer 定义了序列化和反序列化的接口
type Serializer interface {
	// Marshal 将值编码为字节
	Marshal(v any) ([]byte, error)
	// Unmarshal 将字节解码为值
	Unmarshal(data []byte, v any) error
}

// jsonSerializer 使用 JSON 格式进行序列化
type jsonSerializer struct{}

func (s *jsonSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s *jsonSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// gobSerializer 使用 Gob 格式进行序列化
type gobSerializer struct{}

func (s *gobSerializer) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *gobSerializer) Unmarshal(data []byte, v any) error {
	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	return dec.Decode(v)
}

// JSONSerializer 是跨语言兼容的 JSON 序列化器（较慢）
var JSONSerializer Serializer = &jsonSerializer{}

// GobSerializer 是 Go 专用的 Gob 序列化器（较快）
var GobSerializer Serializer = &gobSerializer{}
