package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/fprofit/telegram-captcha-bot/internal/app"
	"github.com/fprofit/telegram-captcha-bot/internal/config"
	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(
		appLog.NewHandler(os.Stdout, level),
	)

	app.Run(cfg, logger)
}
