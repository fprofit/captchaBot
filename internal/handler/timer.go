package handler

import (
	"context"
)

func (h *Handler) StartCaptchaTimer(ctx context.Context) {
	h.logger.Info("CAPTCHA event listener started (zero-ticker mode)")
	expiredChan := h.captcha.Expired()

	for {
		select {
		case challenge := <-expiredChan:
			h.logger.Info("User failed captcha timeout", "user_id", challenge.UserID)

			h.deleteCaptchaMessage(challenge.Key.ChatID, challenge.Key.MessageID)
			h.client.RemoveChatMember(challenge.Key.ChatID, challenge.UserID)
			h.notificationRemovedExpired(challenge.Key.ChatID, challenge.UserID)

		case <-ctx.Done():
			return
		}
	}
}
