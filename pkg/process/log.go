package process

import "log/slog"

type ILogger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

var logger ILogger = slog.Default()

func SetLogger(l ILogger) {
	logger = l
}
