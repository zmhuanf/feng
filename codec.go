package feng

import "github.com/zmhuanf/feng/internal/core"

type Codec = core.Codec

func NewJSONCodec() Codec {
	return core.NewJSONCodec()
}

func NewJsonCodec() Codec {
	return NewJSONCodec()
}
