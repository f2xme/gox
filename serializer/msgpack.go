package serializer

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// MsgPackSerializer MessagePack 序列化器
type MsgPackSerializer struct{}

// NewMsgPack 创建 MessagePack 序列化器
func NewMsgPack() Serializer {
	return &MsgPackSerializer{}
}

func (s *MsgPackSerializer) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (s *MsgPackSerializer) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func (s *MsgPackSerializer) Encode(w io.Writer, v any) error {
	return msgpack.NewEncoder(w).Encode(v)
}

func (s *MsgPackSerializer) Decode(r io.Reader, v any) error {
	return msgpack.NewDecoder(r).Decode(v)
}

func (s *MsgPackSerializer) ContentType() string {
	return "application/msgpack"
}
