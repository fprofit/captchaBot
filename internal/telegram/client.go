package telegram

import (
	"net/http"
	"time"

	applog "github.com/fprofit/telegram-captcha-bot/internal/log"
)

type Client struct {
	token       string
	pollTimeout int
	client      *http.Client
	logger      *applog.Logger
}

func NewClient(token string, pollTimeout int, logger *applog.Logger) *Client {
	return &Client{
		token:       token,
		pollTimeout: pollTimeout,
		client: &http.Client{
			Timeout: 40 * time.Second,
		},
		logger: logger,
	}
}
