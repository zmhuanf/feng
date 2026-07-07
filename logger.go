package feng

import "github.com/zmhuanf/feng/internal/core"

type Logger = core.Logger

func NewSlogLogger() Logger {
	return core.NewSlogLogger()
}
