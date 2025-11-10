package client

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	config := NewDefaultClientConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger

	client := NewClient(config)
	err := client.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	err = client.Request("/test", "hello world", func(ctx IContext, data string) {
		t.Logf("response: %v", data)
	}, time.Second*30)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
}

func TestClient2(t *testing.T) {
	config := NewDefaultClientConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger

	client := NewClient(config)
	err := client.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	err = client.AddHandler("/res_1", func(ctx IContext, data string) {
		t.Logf("response: %v", data)
	})
	if err != nil {
		t.Fatalf("add handler failed: %v", err)
	}
	err = client.Push("/test_1", "hello world!")
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	time.Sleep(time.Second * 5)
}
