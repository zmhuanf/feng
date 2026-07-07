package feng

import "github.com/zmhuanf/feng/internal/core"

type ServerConfig = core.ServerConfig
type ClientConfig = core.ClientConfig
type Mode = core.Mode

const (
	ModeClient = core.ModeClient
	ModeServer = core.ModeServer
)

func NewDefaultServerConfig() ServerConfig {
	return core.NewDefaultServerConfig()
}

func NewDefaultClientConfig() ClientConfig {
	return core.NewDefaultClientConfig()
}
