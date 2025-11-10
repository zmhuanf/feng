package feng

import "testing"

func TestLogger(t *testing.T) {
	log := NewSlogLogger()
	log.Info("test", "data", "hello world")
	log.Error("test", "data", "hello world")
}