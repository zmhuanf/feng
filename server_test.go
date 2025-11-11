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

func TestServer2(t *testing.T) {
	config := NewDefaultServerConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger
	server := NewServer(config)

	err := server.AddHandler("/test_1", func(ctx IContext, data string) (string, error) {
		t.Logf("In TestServer2 /test_1: %v", data)
		err := ctx.GetUser().Push("/res_1", data)
		if err != nil {
			t.Fatalf("push failed: %v", err)
			return "", err
		}
		return data, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	server.Start()
}

func TestServer3(t *testing.T) {
	config := NewDefaultServerConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger
	server := NewServer(config)

	type A struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := server.AddHandler("/test_3", func(ctx IContext, a A) (A, error) {
		t.Logf("In TestServer3 /test_3: %v", a)
		a.Age += 100
		return a, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	server.Start()
}
