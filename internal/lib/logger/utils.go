package logger

import (
	"greenlight/proj/internal/lib/logger/handlers/slogpretty"
	"log"
	"log/slog"
	"os"
)

func SetupLogger(debug bool) *slog.Logger {
	var handler slog.Handler
	if debug {
		handler = slogpretty.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(handler)
}

type out struct {
	stdLog *slog.Logger
}

func (l out) Write(p []byte) (n int, err error) {
	l.stdLog.Info(string(p))
	return len(p), nil
}

func LogAdapter(logger *slog.Logger) *log.Logger {
	return log.New(&out{logger}, "", 0)
}
