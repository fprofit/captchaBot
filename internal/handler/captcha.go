// internal/handler/captcha.go

package handler

import (
	"github.com/fprofit/telegram-captcha-bot/internal/captcha"
	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func (h *Handler) generateCaptcha() (captcha.GeneratedChallenge, *telegram.InlineKeyboardMarkup) {
	challenge, err := h.captcha.Generator.Generate(4)
	if err != nil {
		h.logger.Error("failed to generate CAPTCHA", "error", err)

		return captcha.GeneratedChallenge{}, nil
	}

	buttons := make([]telegram.InlineKeyboardButton, 0, len(challenge.Options))

	for _, option := range challenge.Options {
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
	h.logger.Debug("CAPTCHA generated", "options", len(challenge.Options))
	return challenge, keyboard
}
