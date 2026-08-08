package app

import (
	stdlog "log"

	"github.com/fprofit/telegram-captcha-bot/internal/config"
	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func Run(cfg config.Config) {
	logger, err := appLog.New(cfg.Debug, cfg.LogLevel)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer logger.Close()

	client := telegram.NewClient(
		cfg.BotToken,
		cfg.PollTimeout,
		logger,
	)

	bot, err := client.GetMe()
	if err != nil {
		logger.Errorf("failed to initialize Telegram bot: %v", err)
		return
	}

	logger.Infof(
		"bot started: @%s (id=%d)",
		bot.Username,
		bot.ID,
	)
}
