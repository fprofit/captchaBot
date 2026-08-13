package handler

import (
	"context"
	"time"
)

func (h *Handler) StartCaptchaTimer(ctx context.Context) {
	h.logger.Info("CAPTCHA timer started")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkCaptchaTimeout()

		case <-ctx.Done():
			h.logger.Info("CAPTCHA timer stopped", "reason", ctx.Err())
			return
		}
	}
}

func (h *Handler) checkCaptchaTimeout() {
	expired := h.captcha.Expired()

	if len(expired) == 0 {
		return
	}

	h.logger.Info("expired CAPTCHA challenges found", "count", len(expired))

	for key, challenge := range expired {
		h.notificationRemovedExpired(key.ChatID, challenge.UserID)
		h.logger.Debug("processing expired CAPTCHA", "chat_id", key.ChatID, "user_id", challenge.UserID, "message_id", key.MessageID)
		h.deleteCaptchaMessage(key.ChatID, key.MessageID)

		if err := h.client.RemoveChatMember(key.ChatID, challenge.UserID); err != nil {
			h.logger.Error("failed to remove user after CAPTCHA timeout", "err", err, "chat_id", key.ChatID, "user_id", challenge.UserID)
			continue
		}
		h.logger.Info("user removed because CAPTCHA expired", "chat_id", key.ChatID, "user_id", challenge.UserID, "message_id", key.MessageID)
	}
}
