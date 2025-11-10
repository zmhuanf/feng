package feng

import (
	"log/slog"
	"os"
	"testing"
)

func TestServer(t *testing.T) {
	config := NewDefaultServerConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger
	server := NewServer(config)

	err := server.AddHandler("/test",
		func(ctx IContext, data string) (string, error) {
			slog.Info("test", "data", data)
			return data, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	server.Start()
}
