package main

import (
	stdlog "log"

	"github.com/fprofit/telegram-captcha-bot/internal/app"
	"github.com/fprofit/telegram-captcha-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatal(err)
	}
	app.Run(cfg)
}
