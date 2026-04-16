package serializer

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// ProtobufSerializer Protobuf 序列化器
type ProtobufSerializer struct{}

// NewProtobuf 创建 Protobuf 序列化器
func NewProtobuf() Serializer {
	return &ProtobufSerializer{}
}

func (s *ProtobufSerializer) Marshal(v any) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("value must implement proto.Message")
	}
	return proto.Marshal(msg)
}

func (s *ProtobufSerializer) Unmarshal(data []byte, v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("value must implement proto.Message")
	}
	return proto.Unmarshal(data, msg)
}

func (s *ProtobufSerializer) Encode(w io.Writer, v any) error {
	data, err := s.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (s *ProtobufSerializer) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return s.Unmarshal(data, v)
}

func (s *ProtobufSerializer) ContentType() string {
	return "application/protobuf"
}
