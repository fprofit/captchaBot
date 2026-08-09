package handler

import (
	"context"
	"time"
)

func (h *Handler) StartCaptchaTimer(ctx context.Context) {
	h.logger.Infof("CAPTCHA timer started")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkCaptchaTimeout()

		case <-ctx.Done():
			h.logger.Infof("CAPTCHA timer stopped")
			return
		}
	}
}

func (h *Handler) checkCaptchaTimeout() {
	expired := h.captcha.Expired()

	if len(expired) == 0 {
		return
	}

	h.logger.Infof("expired CAPTCHA challenges found: count=%d", len(expired))

	for key, challenge := range expired {
		h.logger.Infof("processing expired CAPTCHA: chat_id=%d user_id=%d message_id=%d",
			key.ChatID, challenge.UserID, key.MessageID)

		if err := h.client.DeleteMessage(key.ChatID, key.MessageID); err != nil {
			h.logger.Errorf("failed to delete expired CAPTCHA: chat_id=%d message_id=%d error=%v",
				key.ChatID, key.MessageID, err)
		} else {
			h.logger.Infof("expired CAPTCHA message deleted: chat_id=%d message_id=%d",
				key.ChatID, key.MessageID)
		}

		if err := h.client.RemoveChatMember(key.ChatID, challenge.UserID); err != nil {
			h.logger.Errorf("failed to kick user after CAPTCHA timeout: chat_id=%d user_id=%d error=%v",
				key.ChatID, challenge.UserID, err)
			continue
		}

		h.logger.Infof("user kicked because CAPTCHA expired: chat_id=%d user_id=%d message_id=%d",
			key.ChatID, challenge.UserID, key.MessageID)
	}
}
