package feng

import (
	"github.com/zmhuanf/feng/internal/core"
	internalserver "github.com/zmhuanf/feng/internal/server"
)

type Server = core.Server

func NewServer(config ServerConfig) Server {
	return internalserver.New(config)
}
