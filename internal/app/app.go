package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/config"
	"github.com/fprofit/telegram-captcha-bot/internal/handler"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func Run(cfg config.Config, logger *slog.Logger) {
	logger.Info("application starting")

	client := telegram.NewClient(
		cfg.BotToken,
		cfg.PollTimeout,
		logger,
	)

	logger.Info("Telegram client initialized")

	bot, err := client.GetMe()
	if err != nil {
		logger.Error("failed to initialize Telegram bot", "err", err)
		return
	}

	logger.Info(
		"bot started",
		"username", bot.Username,
		"id", bot.ID,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Info("signal handler initialized")

	captchaManager := captcha.NewManager(
		time.Duration(cfg.CaptchaTimeout)*time.Second,
		logger,
	)

	logger.Info("CAPTCHA manager initialized", "timeout", cfg.CaptchaTimeout)

	handler := handler.NewHandler(
		client,
		logger,
		captchaManager,
	)

	logger.Info("handler initialized")

	go handler.StartCaptchaTimer(ctx)

	logger.Info("CAPTCHA timer started")
	logger.Info("starting Telegram polling")

	err = client.Poll(ctx, handler.Handle)
	if err != nil && ctx.Err() == nil {
		logger.Error("polling stopped with error", "err", err)
	}

	logger.Info("Telegram polling stopped")

	if ctx.Err() != nil {
		logger.Info("application stopped by signal")
	}

	logger.Info("application shutdown complete")
}
