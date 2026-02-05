package logger

import (
	"log/slog"
	"os"

	"github.com/nikallow/bookstores-api/internal/config"
)

func SetupLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logger.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch cfg.Env {
	case config.EnvLocal, config.EnvDev:
		handler = slog.NewTextHandler(os.Stdout, opts)
	case config.EnvProd:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}
