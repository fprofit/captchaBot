package handler

import (
	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func (h *Handler) handleCallback(callback *telegram.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID
	messageID := callback.Message.MessageID

	h.logger.Debug(
		"callback received",
		"callback_id", callback.ID,
		"data", callback.Data,
		"user_id", userID,
		"chat_id", chatID,
		"message_id", messageID,
	)

	result := h.captcha.Check(
		captcha.ChallengeKey{
			ChatID:    chatID,
			MessageID: messageID,
		},
		userID,
		callback.Data,
	)

	h.logger.Debug(
		"CAPTCHA check completed",
		"user_id", userID,
		"chat_id", chatID,
		"message_id", messageID,
		"result", result,
	)

	var text string
	showAlert := false

	switch result {
	case 1:
		text = "Добро пожаловать! 🎉\n\nБудьте вежливы 🤝\nНе спамьте 🚫\nУважайте участников."
		showAlert = true

		h.logger.Info(
			"CAPTCHA passed",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", messageID,
		)

		h.deleteCaptchaMessage(chatID, messageID)

		if err := h.client.UnrestrictChatMember(chatID, userID); err != nil {
			h.logger.Error(
				"failed to unrestrict user after CAPTCHA",
				"err", err,
				"user_id", userID,
				"chat_id", chatID,
			)
		} else {
			h.logger.Info(
				"user unrestricted after CAPTCHA",
				"user_id", userID,
				"chat_id", chatID,
			)
		}

	case 2:
		text = "🚫 Эта CAPTCHA не для тебя."

		h.logger.Warn(
			"CAPTCHA belongs to another user",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", messageID,
		)

	case 3:
		text = "❗ CAPTCHA больше недействительна."

		h.logger.Info(
			"CAPTCHA expired",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", messageID,
		)

		if err := h.client.DeleteMessage(chatID, messageID); err != nil {
			h.logger.Error(
				"failed to delete expired CAPTCHA message",
				"err", err,
				"chat_id", chatID,
				"message_id", messageID,
			)
		}

	case 4:
		text = "❗ Попробуй снова"

		answers, err := h.captcha.Generator.Generate(4)
		h.logger.Error(
			"failed to restrict new member",
			"err", err,
			"user_id", userID,
			"chat_id", chatID,
		)

		buttons := make(
			[]telegram.InlineKeyboardButton,
			0,
			len(answers.Options),
		)

		for _, answer := range answers.Options {
			buttons = append(
				buttons,
				telegram.InlineKeyboardButton{
					Text:         answer.Text,
					CallbackData: answer.Token,
				},
			)
		}

		keyboard := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				buttons,
			},
		}

		h.logger.Debug(
			"CAPTCHA answer was incorrect, generating new challenge",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", messageID,
		)

		sentMessage, err := h.client.EditCaptchaMessageImg(
			chatID,
			messageID,
			answers.Image,
			keyboard,
		)
		if err != nil {
			h.logger.Error(
				"failed to update CAPTCHA message",
				"err", err,
				"user_id", userID,
				"chat_id", chatID,
				"message_id", messageID,
			)

			break
		}

		h.captcha.Create(
			captcha.ChallengeKey{
				ChatID:    chatID,
				MessageID: sentMessage.MessageID,
			},
			userID,
			answers,
		)

		h.logger.Debug(
			"new CAPTCHA challenge created",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", sentMessage.MessageID,
		)

	default:
		text = "❌ CAPTCHA неверная."

		h.logger.Info(
			"CAPTCHA failed",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", messageID,
		)
	}

	if err := h.client.AnswerCallbackQuery(
		callback.ID,
		text,
		showAlert,
	); err != nil {
		h.logger.Error(
			"failed to answer callback query",
			"err", err,
			"callback_id", callback.ID,
			"user_id", userID,
			"chat_id", chatID,
		)
	}
}

func (h *Handler) deleteCaptchaMessage(chatID, messageID int64) {
	h.logger.Debug(
		"deleting CAPTCHA message",
		"chat_id", chatID,
		"message_id", messageID,
	)

	h.captcha.Delete(captcha.ChallengeKey{
		ChatID:    chatID,
		MessageID: messageID,
	})

	if err := h.client.DeleteMessage(chatID, messageID); err != nil {
		h.logger.Error(
			"failed to delete CAPTCHA message",
			"err", err,
			"chat_id", chatID,
			"message_id", messageID,
		)

		return
	}

	h.logger.Debug(
		"CAPTCHA message deleted",
		"chat_id", chatID,
		"message_id", messageID,
	)
}
