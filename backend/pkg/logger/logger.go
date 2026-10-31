package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger.
func New(level string, format string) *slog.Logger {
	var opts slog.HandlerOptions
	switch level {
	case "debug":
		opts = slog.HandlerOptions{Level: slog.LevelDebug}
	case "warn":
		opts = slog.HandlerOptions{Level: Warn}
	case "error":
		opts = slog.HandlerOptions{Level: slog.LevelError}
	default:
		opts = slog.HandlerOptions{Level: slog.LevelInfo}
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &opts)
	}

	return slog.New(handler)
}
