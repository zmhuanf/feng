package core

import (
	"bytes"
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/proto"
)

// ErrInvalidProtoMessage 表示传入对象未实现 proto.Message 接口。
var ErrInvalidProtoMessage = errors.New("codec: value does not implement proto.Message")

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type jsonCodec struct{}

func NewJSONCodec() Codec {
	return jsonCodec{}
}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case nil:
		return []byte(""), nil
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case json.RawMessage:
		return val, nil
	case *bytes.Buffer:
		return val.Bytes(), nil
	}
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type protoCodec struct{}

func NewProtoCodec() Codec {
	return protoCodec{}
}

func (protoCodec) Marshal(v any) ([]byte, error) {
	if v == nil {
		return []byte(""), nil
	}
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, ErrInvalidProtoMessage
	}
	return proto.Marshal(msg)
}

func (protoCodec) Unmarshal(data []byte, v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidProtoMessage
	}
	return proto.Unmarshal(data, msg)
}
