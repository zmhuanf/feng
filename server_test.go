package feng

import (
	"log/slog"
	"testing"
)

func TestServer(t *testing.T) {
	server := NewServer(NewDefaultServerConfig())
	server.AddHandler("/test",
		func(ctx IContext, data string) string {
			slog.Info("test", "data", data)
			return data
		},
	)
	server.Start()
}
