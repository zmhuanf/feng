package feng

import (
	"errors"
	"log/slog"
	"os"
	"testing"
)

func TestServer1(t *testing.T) {
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
		func(ctx IServerContext, data string) (string, error) {
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

	err := server.AddHandler("/test_1", func(ctx IServerContext, data string) (string, error) {
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

	err := server.AddHandler("/test_3", func(ctx IServerContext, a A) (A, error) {
		t.Logf("In TestServer3 /test_3: %v", a)
		a.Age += 100
		return a, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	server.Start()
}

func TestServer4(t *testing.T) {
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

	err := server.AddHandler(
		"/cocos_test",
		func(ctx IServerContext, a A) (A, error) {
			t.Logf("In TestServer4 /cocos_test: %v", a)

			err := ctx.GetUser().Push("/hello", a)
			if err != nil {
				t.Fatalf("push failed: %v", err)
				return a, err
			}

			a.Age += 100
			a.Name += "_good"

			return a, errors.New("test error")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	server.Start()
}

func TestServer5(t *testing.T) {
	config := NewDefaultServerConfig()
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	config.Logger = logger
	server := NewServer(config)

	err := server.AddHandler(
		"/test5",
		func(ctx IServerContext, data bool) error {
			t.Logf("In TestServer5 /test5: %v", data)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// err = server.AddHandler(
	// 	"/test5_2",
	// 	func(ctx IContext, data []byte) error {
	// 		t.Logf("In TestServer5 /test5_2: %v", data)
	// 		return nil
	// 	},
	// )
	// if err != nil {
	// 	t.Fatal(err)
	// }
	server.Start()
}
