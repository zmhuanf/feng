package feng

import (
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	client := NewClient(NewDefaultClientConfig())
	err := client.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	err = client.Request("/test", "hello world", func(ctx IContext, data any) {
		t.Logf("response: %v", data)
	}, time.Second*10)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
}
