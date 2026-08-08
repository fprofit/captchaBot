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
	Debug          bool
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

	pollTimeout, err := strconv.Atoi(os.Getenv("POLL_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid POLL_TIMEOUT: %w",
			err,
		)
	}

	captchaTimeout, err := strconv.Atoi(os.Getenv("CAPTCHA_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid CAPTCHA_TIMEOUT: %w",
			err,
		)
	}

	debug := true
	if value := os.Getenv("DEBUG"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			debug = parsed
		}
	}

	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))

	switch logLevel {
	case "DEBUG", "INFO", "WARN", "ERROR":
	default:
		logLevel = "INFO"
	}
	return Config{
		BotToken:       token,
		PollTimeout:    pollTimeout,
		CaptchaTimeout: captchaTimeout,
		Debug:          debug,
		LogLevel:       logLevel,
	}, nil
}
