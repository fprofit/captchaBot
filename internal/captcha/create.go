package captcha

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"time"
)

type AnswerCallbackQuery struct {
	Text         string
	CallbackData string
}

var emojis = []string{
	"😀", "🐶", "🦊", "🐼", "🐸", "🐵", "🍒",
	"🦁", "🐯", "🐨", "🍎", "🍋", "🍉", "🍓",
	"🥝", "🍌", "🚗", "🚕", "🚓", "✈️", "🚀",
	"🏀", "🎲", "⭐", "🌙", "🔥", "❄️", "⚽",
}

func (m *Manager) CreateAnswers(
	chatID int64,
	messageID int64,
	count int,
) []AnswerCallbackQuery {
	m.logger.Debugf(
		"creating CAPTCHA answers: chat_id=%d message_id=%d count=%d",
		chatID,
		messageID,
		count,
	)

	emojis := randomEmojis(count)

	if len(emojis) == 0 {
		m.logger.Warnf(
			"failed to generate CAPTCHA answers: chat_id=%d message_id=%d count=%d",
			chatID,
			messageID,
			count,
		)
		return nil
	}

	answers := make([]AnswerCallbackQuery, len(emojis))
	t := time.Now().UnixMicro()

	for i, emoji := range emojis {
		data := fmt.Sprintf(
			"%d:%d:%s:%d",
			chatID,
			messageID,
			emoji,
			t,
		)

		hash := sha256.Sum256([]byte(data))

		answers[i] = AnswerCallbackQuery{
			Text:         emoji,
			CallbackData: hex.EncodeToString(hash[:16]),
		}
	}

	m.logger.Debugf(
		"CAPTCHA answers created: chat_id=%d message_id=%d count=%d",
		chatID,
		messageID,
		len(answers),
	)

	return answers
}

func randomEmojis(count int) []string {
	if count > len(emojis) {
		count = len(emojis)
	}

	result := make([]string, len(emojis))
	copy(result, emojis)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result[:count]
}
