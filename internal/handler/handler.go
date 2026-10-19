package handler

import (
	"log/slog"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

type Handler struct {
	client      *telegram.Client
	adminUserID int64
	logger      *slog.Logger
	captcha     *captcha.Manager
}

func NewHandler(client *telegram.Client, userID int64, logger *slog.Logger, captcha *captcha.Manager) *Handler {
	logger.Debug("creating handler")

	return &Handler{
		client:      client,
		logger:      logger,
		adminUserID: userID,
		captcha:     captcha,
	}
}

func (h *Handler) Handle(update telegram.Update) {
	h.logger.Debug("handling Telegram update", "update_id", update.UpdateID)

	if update.CallbackQuery != nil {
		h.logger.Debug("update contains callback query", "update_id", update.UpdateID, "callback_id", update.CallbackQuery.ID)
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.ChatMember != nil {
		cm := update.ChatMember

		isJoining := cm.NewChatMember.Status == "member" &&
			(cm.OldChatMember.Status == "left" || cm.OldChatMember.Status == "kicked")

		if isJoining {
			h.logger.Info("new user joined the chat, triggering CAPTCHA",
				"chat_id", cm.Chat.ID,
				"user_id", cm.NewChatMember.User.ID,
			)
			h.sendCaptchaNewUser(cm.Chat, cm.NewChatMember.User)
		}
		return
	}

	if update.Message != nil {
		if len(update.Message.NewChatMembers) > 0 || update.Message.LeftChatMember != nil {
			h.logger.Info("deleting service message", "chat_id", update.Message.Chat.ID, "message_id", update.Message.MessageID)
			if err := h.client.DeleteMessage(update.Message.Chat.ID, update.Message.MessageID); err != nil {
				h.logger.Error("failed to delete service message", "err", err, "chat_id", update.Message.Chat.ID, "message_id", update.Message.MessageID)
			}
			return
		}

		h.logger.Debug("ignoring standard message layout", "message_id", update.Message.MessageID)
		return
	}
}
