package logger

import (
	"log/slog"
	"os"
)

// Init sets up the global logger with JSON handler.
func Init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

