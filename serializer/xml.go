package serializer

import (
	"encoding/xml"
	"io"
)

// XMLSerializer XML 序列化器
type XMLSerializer struct{}

// NewXML 创建 XML 序列化器
func NewXML() Serializer {
	return &XMLSerializer{}
}

func (s *XMLSerializer) Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}

func (s *XMLSerializer) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

func (s *XMLSerializer) Encode(w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (s *XMLSerializer) Decode(r io.Reader, v any) error {
	return xml.NewDecoder(r).Decode(v)
}

func (s *XMLSerializer) ContentType() string {
	return "application/xml"
}
