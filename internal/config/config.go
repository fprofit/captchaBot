package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fprofit/telegram-captcha-bot/internal/env"
)

type Config struct {
	BotToken       string
	PollTimeout    int
	CaptchaTimeout int
	AdminUserID    int64
	LogLevel       string
}

func Load() (Config, error) {
	if err := env.Load(); err != nil {
		return Config{}, fmt.Errorf("load env: %w", err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("BOT_TOKEN is not set")
	}

	adminUserID, err := strconv.ParseInt(os.Getenv("ADMIN_USER_ID"), 10, 64)
	if err != nil || adminUserID == 0 {
		adminUserID = -1
	}

	pollTimeout, err := strconv.Atoi(os.Getenv("POLL_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid POLL_TIMEOUT: %w",
			err,
		)
	}
	if pollTimeout < 0 {
		return Config{}, fmt.Errorf(
			"POLL_TIMEOUT must be greater than zero",
		)
	}

	captchaTimeout, err := strconv.Atoi(os.Getenv("CAPTCHA_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid CAPTCHA_TIMEOUT: %w",
			err,
		)
	}
	if captchaTimeout <= 0 {
		captchaTimeout = 300
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch logLevel {
	case "debug", "info", "warn", "error":
	default:
		logLevel = "info"
	}
	return Config{
		BotToken:       token,
		AdminUserID:    adminUserID,
		PollTimeout:    pollTimeout,
		CaptchaTimeout: captchaTimeout,
		LogLevel:       logLevel,
	}, nil
}
