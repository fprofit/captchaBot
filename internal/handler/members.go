package handler

import (
	"fmt"

	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

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

		answers, err := h.captcha.Generator.Generate(4)
		h.logger.Error("failed to restrict new member", "err", err, "user_id", userID, "chat_id", chatID)

		h.logger.Debug("CAPTCHA challenge generated", "user_id", userID, "chat_id", chatID, "answers", len(answers.Options))

		if len(answers.Options) == 0 {
			h.logger.Error("failed to generate CAPTCHA answers", "user_id", userID, "chat_id", chatID)
			continue
		}

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

		h.logger.Debug("CAPTCHA keyboard created",
			"user_id", userID,
			"chat_id", chatID,
			"buttons", len(buttons),
		)

		sentMessage, err := h.client.SendCaptchaMessageImg(chatID,
			fmt.Sprintf("Привет, <a href=\"tg://user?id=%d\">%s %s</a>!\n",
				userID,
				user.FirstName,
				user.LastName,
			),
			answers.Image,
			keyboard,
		)
		if err != nil {
			h.logger.Error("failed to send CAPTCHA message",
				"err", err,
				"user_id", userID,
				"chat_id", chatID,
			)

			continue
		}

		h.logger.Debug(
			"CAPTCHA message sent",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", sentMessage.MessageID,
		)

		h.captcha.Create(
			captcha.ChallengeKey{
				ChatID:    chatID,
				MessageID: sentMessage.MessageID,
			},
			userID,
			answers,
		)

		h.logger.Info("CAPTCHA created",
			"user_id", userID,
			"chat_id", chatID,
			"message_id", sentMessage.MessageID,
		)
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
