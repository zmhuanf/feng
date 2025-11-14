package client

import (
	"context"
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
	err = client.Request(context.Background(), "/test", "hello world", func(ctx IContext, data string) {
		t.Logf("response: %v", data)
	})
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

func TestClient3(t *testing.T) {
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

	type A struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err = client.Request(context.Background(), "/test_3",
		A{Name: "feng", Age: 18},
		func(ctx IContext, a A) {
			t.Logf("response: %v", a)
		},
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	time.Sleep(time.Second * 5)
}

func TestClient5(t *testing.T) {
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

	err = client.Request(
		context.Background(),
		"/test5",
		true,
		func(ctx IContext, data string) {
			t.Logf("response: %v", data)
		},
	)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	// err = client.Push(
	// 	"/test5_2",
	// 	"",
	// )
	// if err != nil {
	// 	t.Fatalf("push failed: %v", err)
	// }
}
