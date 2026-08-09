package handler

import "github.com/fprofit/telegram-captcha-bot/internal/telegram"

func (h *Handler) handleCallback(callback *telegram.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID
	messageID := callback.Message.MessageID

	h.logger.Infof("callback received: id=%s data=%s user_id=%d chat_id=%d message_id=%d",
		callback.ID, callback.Data, userID, chatID, messageID,
	)

	result := h.captcha.Check(chatID, messageID, userID, callback.Data)

	h.logger.Debugf("CAPTCHA check result: user_id=%d chat_id=%d message_id=%d result=%d",
		userID, chatID, messageID, result,
	)

	var text string
	showAlert := false

	switch result {
	case 1:
		text = "✅ CAPTCHA пройдена!\nПравила: уважайте других участников и соблюдайте правила группы."
		showAlert = true

		h.logger.Infof("CAPTCHA passed: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, messageID,
		)

		h.deleteCaptchaMessage(chatID, messageID)

		if err := h.client.UnrestrictChatMember(chatID, userID); err != nil {
			h.logger.Errorf("failed to unrestrict user: user_id=%d chat_id=%d error=%v",
				userID, chatID, err,
			)
		} else {
			h.logger.Infof("user unrestricted after CAPTCHA: user_id=%d chat_id=%d",
				userID, chatID,
			)
		}

	case 2:
		text = "🚫 Эта CAPTCHA не для тебя."

		h.logger.Warnf("CAPTCHA belongs to another user: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, messageID)

	case 3:
		text = "❗ CAPTCHA больше недействительна."

		h.logger.Infof("CAPTCHA expired: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, messageID,
		)

		if err := h.client.DeleteMessage(chatID, messageID); err != nil {
			h.logger.Errorf("failed to delete expired CAPTCHA message: chat_id=%d message_id=%d error=%v",
				chatID, messageID, err,
			)
		}

	default:
		text = "❌ CAPTCHA неверная."

		h.logger.Infof("CAPTCHA failed: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, messageID)
	}

	if err := h.client.AnswerCallbackQuery(
		callback.ID, text, showAlert); err != nil {
		h.logger.Errorf("failed to answer callback query: %v", err)
	}
}

func (h *Handler) deleteCaptchaMessage(chatID, messageID int64) {
	h.logger.Debugf("deleting CAPTCHA: chat_id=%d message_id=%d",
		chatID, messageID)

	h.captcha.Delete(chatID, messageID)

	if err := h.client.DeleteMessage(chatID, messageID); err != nil {
		h.logger.Errorf("failed to delete CAPTCHA message: chat_id=%d message_id=%d error=%v",
			chatID, messageID, err,
		)
		return
	}

	h.logger.Infof("CAPTCHA message deleted: chat_id=%d message_id=%d",
		chatID, messageID,
	)
}
