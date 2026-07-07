package core

import (
	"bytes"
	"encoding/json"
)

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
