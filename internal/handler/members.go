package handler

import (
	"fmt"
	"html"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func formatCaptchaCaption(user telegram.User) string {
	return fmt.Sprintf(
		"Привет, <a href=\"tg://user?id=%d\">%s %s</a>! 👋\n\nКто не решил - тот бот 🤖",
		user.ID,
		html.EscapeString(user.FirstName),
		html.EscapeString(user.LastName),
	)
}

func (h *Handler) handleNewChatMembers(message *telegram.Message) {
	chatID := message.Chat.ID

	for _, user := range message.NewChatMembers {
		userID := user.ID

		h.logger.Info("new member joined", "user_id", userID, "username", user.Username, "chat_id", chatID)

		if err := h.client.DeleteMessage(chatID, message.MessageID); err != nil {
			h.logger.Error("failed to delete service message", "err", err, "chat_id", chatID, "message_id", message.MessageID)
		} else {
			h.logger.Debug("service message deleted", "chat_id", chatID, "message_id", message.MessageID)
		}

		if err := h.client.RestrictChatMember(chatID, userID); err != nil {
			h.logger.Error("failed to restrict new member", "err", err, "user_id", userID, "chat_id", chatID)
			continue
		}

		h.logger.Info("user restricted until CAPTCHA completion", "user_id", userID, "chat_id", chatID)

		g, k := h.generateCaptcha()

		sentMessage, err := h.client.SendCaptchaMessageImg(chatID, formatCaptchaCaption(user), g.Image, k)
		if err != nil {
			h.logger.Error("failed to send CAPTCHA message", "err", err, "user_id", userID, "chat_id", chatID)

			continue
		}

		h.logger.Debug("CAPTCHA message sent", "user_id", userID, "chat_id", chatID, "message_id", sentMessage.MessageID)

		h.captcha.Create(captcha.ChallengeKey{ChatID: chatID, MessageID: sentMessage.MessageID}, userID, g)

		h.logger.Info("CAPTCHA created", "user_id", userID, "chat_id", chatID, "message_id", sentMessage.MessageID)
	}
}

func (h *Handler) handleLeftChatMember(message *telegram.Message) {
	user := message.LeftChatMember
	h.logger.Info("member left", "user_id", user.ID, "chat_id", message.Chat.ID, "message_id", message.MessageID)
	if err := h.client.DeleteMessage(message.Chat.ID, message.MessageID); err != nil {
		h.logger.Error("failed to delete leave message", "err", err, "chat_id", message.Chat.ID, "message_id", message.MessageID)
		return
	}
	h.logger.Debug("leave message deleted", "chat_id", message.Chat.ID, "message_id", message.MessageID)
}
