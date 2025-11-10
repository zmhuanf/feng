package feng

import (
	"bytes"
	"encoding/json"
)

type ICodec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type jsonCodec struct {
}

func (j *jsonCodec) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case nil:
		return []byte("null"), nil
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

func (j *jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewJsonCodec() ICodec {
	return &jsonCodec{}
}
