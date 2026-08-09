package app

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/config"
	"github.com/fprofit/telegram-captcha-bot/internal/handler"
	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func Run(cfg config.Config) {
	logger, err := appLog.New(cfg.Debug, cfg.LogLevel)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer logger.Close()

	logger.Infof("application starting")

	client := telegram.NewClient(
		cfg.BotToken,
		cfg.PollTimeout,
		logger,
	)

	logger.Infof("Telegram client initialized")

	bot, err := client.GetMe()
	if err != nil {
		logger.Errorf("failed to initialize Telegram bot: %v", err)
		return
	}

	logger.Infof("bot started: @%s (id=%d)", bot.Username, bot.ID)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Infof("signal handler initialized")

	captchaManager := captcha.NewManager(cfg.CaptchaTimeout, logger)

	logger.Infof(
		"CAPTCHA manager initialized: timeout=%s",
		cfg.CaptchaTimeout,
	)

	handler := handler.NewHandler(client, logger, captchaManager)

	logger.Infof("handler initialized")

	go handler.StartCaptchaTimer(ctx)

	logger.Infof("CAPTCHA timer started")

	logger.Infof("starting Telegram polling")

	err = client.Poll(ctx, handler.Handle)
	if err != nil && ctx.Err() == nil {
		logger.Errorf("polling stopped with error: %v", err)
	}

	logger.Infof("Telegram polling stopped")

	if ctx.Err() != nil {
		logger.Infof("application stopped by signal")
	}

	logger.Infof("application shutdown complete")
}
