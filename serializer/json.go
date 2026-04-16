package serializer

import (
	"encoding/json"
	"io"
)

// JSONSerializer JSON 序列化器
type JSONSerializer struct{}

// NewJSON 创建 JSON 序列化器
func NewJSON() Serializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (s *JSONSerializer) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (s *JSONSerializer) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (s *JSONSerializer) ContentType() string {
	return "application/json"
}
