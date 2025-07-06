package logging

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Setup(env string) (*slog.Logger, error) {
	var lg *slog.Logger

	switch env {
	case envLocal:
		lg = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))

	case envDev:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))

	default:
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	return lg, nil
}
