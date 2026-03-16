package logger

import (
	"log/slog"
	"os"
)

// New creates a structured JSON logger for production or a human-readable
// text logger for development, based on the APP_ENV environment variable.
func New(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
