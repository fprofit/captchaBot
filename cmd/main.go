package main

import (
	"log/slog"
	"os"

	"github.com/fprofit/telegram-captcha-bot/internal/app"
	"github.com/fprofit/telegram-captcha-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"failed to load configuration",
			"err", err,
		)
		os.Exit(1)
	}

	level := slog.LevelInfo

	level = slog.LevelDebug

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: level,
			},
		),
	)

	app.Run(cfg, logger)
}
