package feng

import (
	"github.com/zmhuanf/feng/internal/client"
	"github.com/zmhuanf/feng/internal/core"
)

type Client = core.Client

func NewClient(config ClientConfig) Client {
	return client.New(config)
}
