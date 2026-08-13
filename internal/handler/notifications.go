package handler

import (
	"fmt"
	"html"

	"github.com/fprofit/telegram-captcha-bot/internal/telegram"
)

func (h *Handler) notification(user telegram.User, chat telegram.Chat, text string) {
	if h.adminUserID == -1 {
		return
	}
	message := fmt.Sprintf(
		"👤 <b>Пользователь</b>\n"+
			"Имя: <a href=\"tg://user?id=%d\">%s %s</a>\n"+
			"Username: %s\n"+
			"ID: <code>%d</code>\n\n"+
			"💬 <b>Чат</b>\n"+
			"Название: %s\n"+
			"Username: %s\n"+
			"ID: <code>%d</code>\n\n"+
			"%s",
		user.ID,
		html.EscapeString(user.FirstName),
		html.EscapeString(user.LastName),
		formatUsername(user.Username),
		user.ID,
		html.EscapeString(chat.Title),
		formatUsername(chat.Username),
		chat.ID,
		text,
	)
	h.client.SendMessage(h.adminUserID, message)
}

func formatUsername(username string) string {
	if username == "" {
		return "—"
	}

	return "@" + html.EscapeString(username)
}

func (h *Handler) notificationRemovedExpired(chatID, userID int64) {
	if h.adminUserID == -1 {
		return
	}
	message := fmt.Sprintf(
		"👤 <b>Пользователь</b>\n"+
			"ID: <code>%d</code>\n\n"+
			"💬 <b>Чат</b>\n"+
			"ID: <code>%d</code>\n\n"+
			"%s",
		userID,
		chatID,
		"Время вышло",
	)
	h.client.SendMessage(h.adminUserID, message)
}
