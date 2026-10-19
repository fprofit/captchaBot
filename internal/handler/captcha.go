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

func (h *Handler) sendCaptchaNewUser(chat telegram.Chat, user telegram.User) {
	userID := user.ID

	h.logger.Info("new member joined", "user_id", userID, "username", user.Username, "chat_id", chat.ID)

	if err := h.client.RestrictChatMember(chat.ID, userID); err != nil {
		h.logger.Error("failed to restrict new member", "err", err, "user_id", userID, "chat_id", chat.ID)
		return
	}

	h.logger.Info("user restricted until CAPTCHA completion", "user_id", userID, "chat_id", chat.ID)

	c, err := captcha.GenerateCaptcha()
	if err != nil {
		h.logger.Error("failed to generate captcha", "err", err)
		return
	}
	k := generateButtons(c.Options)

	sentMessage, err := h.client.SendCaptchaMessageImg(chat.ID, formatCaptchaCaption(user), c.Image, k)
	if err != nil {
		h.logger.Error("failed to send CAPTCHA message", "err", err, "user_id", userID, "chat_id", chat.ID)
		return
	}
	h.captcha.Put(captcha.ChallengeKey{
		ChatID:    chat.ID,
		MessageID: sentMessage.MessageID,
	},
		userID,
		c,
	)

	h.logger.Debug("CAPTCHA message sent", "user_id", userID, "chat_id", chat.ID, "message_id", sentMessage.MessageID)

	h.logger.Info("CAPTCHA created", "user_id", userID, "chat_id", chat.ID, "message_id", sentMessage.MessageID)
}

func (h *Handler) retryCaptcha(chatID, messageID int64, user telegram.User) {
	c, err := captcha.GenerateCaptcha()
	if err != nil {
		h.logger.Error("failed to generate captcha", "err", err)
		return
	}
	k := generateButtons(c.Options)

	sentMessage, err := h.client.EditCaptchaMessageImg(chatID, messageID, formatCaptchaCaption(user), c.Image, k)
	if err != nil {
		h.logger.Error("failed to update CAPTCHA message", "err", err, "user_id", user.ID, "chat_id", chatID, "message_id", messageID)
		return
	}

	h.captcha.Put(captcha.ChallengeKey{
		ChatID:    chatID,
		MessageID: sentMessage.MessageID,
	},
		user.ID,
		c,
	)

	h.logger.Debug("new CAPTCHA challenge created", "user_id", user.ID, "chat_id", chatID, "message_id", sentMessage.MessageID)
}

func generateButtons(k []captcha.Option) *telegram.InlineKeyboardMarkup {
	buttons := make([]telegram.InlineKeyboardButton, 0, len(k))

	for _, option := range k {
		buttons = append(
			buttons,
			telegram.InlineKeyboardButton{
				Text:         option.Text,
				CallbackData: option.Token,
			},
		)
	}

	keyboard := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			buttons,
		},
	}
	return keyboard
}
