package handler

import (
	"time"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

const (
	CaptchaPassedMessage    = "🎉 Добро пожаловать!\n\n🤝 Будьте вежливы\n🚫 Не спамьте\n❤️ Уважайте участников."
	CaptchaWrongUserMessage = "❗ Эта CAPTCHA не для тебя."
	CaptchaExpiredMessage   = "🚫 Эта CAPTCHA больше недействительна."
	CaptchaRetryMessage     = "🔄 Попробуй снова."
	CaptchaKickMessage      = "🚪 Выйди и зайди нормально 😄"
	CaptchaIncorrectMessage = "❌ CAPTCHA неверная."
)

func (h *Handler) handleCallback(callback *telegram.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID
	messageID := callback.Message.MessageID

	h.logger.Debug("callback received", "callback_id", callback.ID, "data", callback.Data, "user_id", userID, "chat_id", chatID, "message_id", messageID)

	result := h.captcha.Check(captcha.ChallengeKey{ChatID: chatID, MessageID: messageID}, userID, callback.Data)

	h.logger.Debug("CAPTCHA check completed", "user_id", userID, "chat_id", chatID, "message_id", messageID, "result", result)

	var text string
	showAlert := false

	switch result {
	case 1:
		text = CaptchaPassedMessage
		h.notification(callback.From, callback.Message.Chat, "✅ Пользователь прошёл CAPTCHA")
		h.captcha.Delete(captcha.ChallengeKey{
			ChatID:    chatID,
			MessageID: messageID,
		})

		showAlert = true
		h.logger.Info("CAPTCHA passed", "user_id", userID, "chat_id", chatID, "message_id", messageID)
		h.deleteCaptchaMessage(chatID, messageID)

		if err := h.client.UnrestrictChatMember(chatID, userID); err != nil {
			h.logger.Error("failed to unrestrict user after CAPTCHA", "err", err, "user_id", userID, "chat_id", chatID)
		} else {
			h.logger.Info("user unrestricted after CAPTCHA", "user_id", userID, "chat_id", chatID)
		}

	case 2:
		text = CaptchaWrongUserMessage
		h.logger.Warn("CAPTCHA belongs to another user", "user_id", userID, "chat_id", chatID, "message_id", messageID)

	case 3:
		text = CaptchaExpiredMessage
		h.logger.Info("CAPTCHA expired", "user_id", userID, "chat_id", chatID, "message_id", messageID)
		h.deleteCaptchaMessage(chatID, messageID)

	case 4:
		text = CaptchaRetryMessage
		h.retryCaptcha(chatID, messageID, callback.From)
	case 5:
		text = CaptchaKickMessage
		h.notification(callback.From, callback.Message.Chat, text)
		showAlert = true
		go h.removeChatMember(chatID, messageID, userID)

	default:
		text = CaptchaIncorrectMessage
		h.logger.Info("CAPTCHA failed", "user_id", userID, "chat_id", chatID, "message_id", messageID)
	}

	if err := h.client.AnswerCallbackQuery(callback.ID, text, showAlert, 3); err != nil {
		h.logger.Error("failed to answer callback query", "err", err, "callback_id", callback.ID, "user_id", userID, "chat_id", chatID)
	}
}

func (h *Handler) deleteCaptchaMessage(chatID, messageID int64) {
	h.logger.Debug("deleting CAPTCHA message", "chat_id", chatID, "message_id", messageID)

	h.captcha.Delete(captcha.ChallengeKey{
		ChatID:    chatID,
		MessageID: messageID,
	})

	if err := h.client.DeleteMessage(chatID, messageID); err != nil {
		h.logger.Error("failed to delete CAPTCHA message", "err", err, "chat_id", chatID, "message_id", messageID)
		return
	}

	h.logger.Debug("CAPTCHA message deleted", "chat_id", chatID, "message_id", messageID)
}

func (h *Handler) removeChatMember(chatID, messageID, userID int64) {
	h.deleteCaptchaMessage(chatID, messageID)
	time.Sleep(5 * time.Second)
	if err := h.client.RemoveChatMember(chatID, userID); err != nil {
		h.logger.Error("failed to remove user after CAPTCHA timeout", "err", err, "chat_id", chatID, "user_id", userID)
	}
}
