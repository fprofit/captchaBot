package handler

import (
	"fmt"
	"math/rand/v2"

	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func (h *Handler) handleNewChatMembers(message *telegram.Message) {
	for _, user := range message.NewChatMembers {
		chatID := message.Chat.ID
		userID := user.ID

		h.logger.Infof("new member: user_id=%d username=@%s chat_id=%d",
			userID, user.Username, chatID,
		)

		if err := h.client.DeleteMessage(chatID, message.MessageID); err != nil {
			h.logger.Errorf("failed to delete service message: chat_id=%d message_id=%d error=%v",
				chatID, message.MessageID, err,
			)
		} else {
			h.logger.Debugf("service message deleted: chat_id=%d message_id=%d",
				chatID, message.MessageID,
			)
		}

		if err := h.client.RestrictChatMember(chatID, userID); err != nil {
			h.logger.Errorf("failed to restrict user: chat_id=%d user_id=%d error=%v",
				chatID, userID, err,
			)
			continue
		}

		h.logger.Infof("user restricted until CAPTCHA completion: chat_id=%d user_id=%d",
			chatID, userID,
		)

		answers := h.captcha.CreateAnswers(chatID, message.MessageID, 4)

		h.logger.Debugf("CAPTCHA answers generated: user_id=%d chat_id=%d count=%d",
			userID, chatID, len(answers),
		)

		if len(answers) == 0 {
			h.logger.Errorf("failed to create CAPTCHA answers: user_id=%d chat_id=%d",
				userID, chatID,
			)
			continue
		}

		buttons := make([]telegram.InlineKeyboardButton, 0, len(answers))

		for _, answer := range answers {
			buttons = append(
				buttons,
				telegram.InlineKeyboardButton{
					Text:         answer.Text,
					CallbackData: answer.CallbackData,
				},
			)
		}

		keyboard := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				buttons,
			},
		}

		correct := answers[rand.IntN(len(answers))]

		h.logger.Debugf("CAPTCHA correct answer selected: user_id=%d chat_id=%d answer=%s",
			userID, chatID, correct.Text,
		)

		sentMessage, err := h.client.SendCaptchaMessage(
			chatID,
			fmt.Sprintf("Привет, <a href=\"tg://user?id=%d\">%s %s</a>!\nНажми на %s",
				userID, user.FirstName, user.LastName, correct.Text,
			),
			keyboard,
		)
		if err != nil {
			h.logger.Errorf("failed to send CAPTCHA message: user_id=%d chat_id=%d error=%v",
				userID, chatID, err,
			)
			continue
		}

		h.logger.Debugf("CAPTCHA message sent: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, sentMessage.MessageID,
		)

		h.captcha.Create(chatID, sentMessage.MessageID, userID, correct)

		h.logger.Infof("CAPTCHA created: user_id=%d chat_id=%d message_id=%d",
			userID, chatID, sentMessage.MessageID,
		)
	}
}

func (h *Handler) handleLeftChatMember(message *telegram.Message) {
	user := message.LeftChatMember

	h.logger.Infof("member left: user_id=%d chat_id=%d message_id=%d",
		user.ID, message.Chat.ID, message.MessageID,
	)

	if err := h.client.DeleteMessage(message.Chat.ID, message.MessageID); err != nil {
		h.logger.Errorf("failed to delete leave message: chat_id=%d message_id=%d error=%v",
			message.Chat.ID, message.MessageID, err,
		)
		return
	}

	h.logger.Infof("leave message deleted: chat_id=%d message_id=%d",
		message.Chat.ID, message.MessageID,
	)
}
