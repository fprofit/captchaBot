package handler

import (
	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

type Handler struct {
	client  *telegram.Client
	logger  *appLog.Logger
	captcha *captcha.Manager
}

func NewHandler(client *telegram.Client, logger *appLog.Logger, captcha *captcha.Manager) *Handler {
	logger.Infof("creating handler")

	return &Handler{
		client:  client,
		logger:  logger,
		captcha: captcha,
	}
}

func (h *Handler) Handle(update telegram.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	if len(update.Message.NewChatMembers) > 0 {
		h.handleNewChatMembers(update.Message)
		return
	}

	if update.Message.LeftChatMember != nil {
		h.handleLeftChatMember(update.Message)
		return
	}
}
