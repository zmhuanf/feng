package feng

import "encoding/json"

type ICodec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type jsonCodec struct {
}

func (j *jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (j *jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewJsonCodec() ICodec {
	return &jsonCodec{}
}
