package captcha

import (
	"sync"
	"time"

	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
)

type Challenge struct {
	UserID    int64
	Correct   AnswerCallbackQuery
	ExpiresAt time.Time
}

type ChallengeKey struct {
	ChatID    int64
	MessageID int64
}

type Manager struct {
	timeout    time.Duration
	mu         sync.RWMutex
	challenges map[ChallengeKey]Challenge
	logger     *appLog.Logger
}

func NewManager(timeout int, logger *appLog.Logger) *Manager {
	logger.Infof("creating CAPTCHA manager: timeout=%s",
		timeout,
	)

	return &Manager{
		timeout:    time.Duration(timeout) * time.Second,
		challenges: make(map[ChallengeKey]Challenge),
		logger:     logger,
	}
}

func (m *Manager) Create(chatID, messageID, userID int64, correct AnswerCallbackQuery) {
	key := ChallengeKey{
		ChatID:    chatID,
		MessageID: messageID,
	}

	challenge := Challenge{
		UserID:    userID,
		Correct:   correct,
		ExpiresAt: time.Now().Add(m.timeout),
	}

	m.mu.Lock()
	m.challenges[key] = challenge
	m.mu.Unlock()

	m.logger.Infof("CAPTCHA created: chat_id=%d user_id=%d message_id=%d correct=%s expires_at=%s",
		chatID, userID, messageID, correct.Text, challenge.ExpiresAt.Format(time.RFC3339),
	)
}

func (m *Manager) Check(chatID, messageID, userID int64, callbackData string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := ChallengeKey{
		ChatID:    chatID,
		MessageID: messageID,
	}

	challenge, ok := m.challenges[key]
	if !ok {
		m.logger.Warnf("CAPTCHA not found: chat_id=%d message_id=%d user_id=%d",
			chatID, messageID, userID,
		)

		return 3
	}

	if challenge.UserID != userID {
		m.logger.Warnf("CAPTCHA belongs to another user: chat_id=%d message_id=%d user_id=%d owner_id=%d",
			chatID, messageID, userID, challenge.UserID,
		)

		return 2
	}

	if challenge.Correct.CallbackData == callbackData {
		m.logger.Infof("CAPTCHA answer correct: chat_id=%d message_id=%d user_id=%d",
			chatID, messageID, userID,
		)

		return 1
	}

	m.logger.Infof("CAPTCHA answer incorrect: chat_id=%d message_id=%d user_id=%d",
		chatID, messageID, userID,
	)

	return 0
}

func (m *Manager) Delete(chatID, messageID int64) {
	key := ChallengeKey{
		ChatID:    chatID,
		MessageID: messageID,
	}

	m.mu.Lock()

	_, exists := m.challenges[key]
	if exists {
		delete(m.challenges, key)
	}

	m.mu.Unlock()

	if !exists {
		m.logger.Debugf("CAPTCHA already deleted or not found: chat_id=%d message_id=%d",
			chatID, messageID,
		)
		return
	}

	m.logger.Infof("CAPTCHA deleted: chat_id=%d message_id=%d",
		chatID, messageID)
}

func (m *Manager) Expired() map[ChallengeKey]Challenge {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expired := make(map[ChallengeKey]Challenge)

	for key, challenge := range m.challenges {
		if now.Before(challenge.ExpiresAt) {
			continue
		}

		expired[key] = challenge
		delete(m.challenges, key)

		m.logger.Infof("CAPTCHA expired: chat_id=%d message_id=%d user_id=%d expires_at=%s",
			key.ChatID, key.MessageID, challenge.UserID, challenge.ExpiresAt.Format(time.RFC3339),
		)
	}

	if len(expired) > 0 {
		m.logger.Infof("expired CAPTCHA challenges removed: count=%d",
			len(expired),
		)
	}

	return expired
}
