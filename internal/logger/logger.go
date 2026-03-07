package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init(env string) {
	var level slog.Level
	switch env {
	case "production":
		level = slog.LevelInfo
	case "development":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	Log = slog.New(handler)
}
