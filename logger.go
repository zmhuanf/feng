package feng

import "log/slog"

type Logger interface {
	Debug(string, ...any)
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

func NewSlogLogger() Logger {
	return slog.Default()
}
