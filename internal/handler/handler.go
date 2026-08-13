package handler

import (
	"log/slog"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

type Handler struct {
	client  *telegram.Client
	logger  *slog.Logger
	captcha *captcha.Manager
}

func NewHandler(client *telegram.Client, logger *slog.Logger, captcha *captcha.Manager) *Handler {
	logger.Debug("creating handler")

	return &Handler{
		client:  client,
		logger:  logger,
		captcha: captcha,
	}
}

func (h *Handler) Handle(update telegram.Update) {
	h.logger.Debug(
		"handling Telegram update",
		"update_id", update.UpdateID,
	)

	if update.CallbackQuery != nil {
		h.logger.Debug(
			"update contains callback query",
			"update_id", update.UpdateID,
			"callback_id", update.CallbackQuery.ID,
		)

		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		h.logger.Debug(
			"ignoring update without message or callback query",
			"update_id", update.UpdateID,
		)

		return
	}

	if len(update.Message.NewChatMembers) > 0 {
		h.logger.Debug(
			"update contains new chat members",
			"update_id", update.UpdateID,
			"chat_id", update.Message.Chat.ID,
			"members", len(update.Message.NewChatMembers),
		)

		h.handleNewChatMembers(update.Message)
		return
	}

	if update.Message.LeftChatMember != nil {
		h.logger.Debug(
			"update contains left chat member",
			"update_id", update.UpdateID,
			"chat_id", update.Message.Chat.ID,
			"user_id", update.Message.LeftChatMember.ID,
		)

		h.handleLeftChatMember(update.Message)
		return
	}

	h.logger.Debug(
		"ignoring unsupported Telegram update",
		"update_id", update.UpdateID,
		"chat_id", update.Message.Chat.ID,
	)
}
