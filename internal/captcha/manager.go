package captcha

import (
	"log/slog"
	"sync"
	"time"
)

type Manager struct {
	Generator  *Generator
	timeout    time.Duration
	mu         sync.RWMutex
	challenges map[ChallengeKey]Challenge
	logger     *slog.Logger
}

func NewManager(timeout time.Duration, logger *slog.Logger) *Manager {
	logger.Debug("creating CAPTCHA manager", "timeout", timeout)

	return &Manager{
		Generator:  NewGenerator(),
		timeout:    timeout,
		challenges: make(map[ChallengeKey]Challenge),
		logger:     logger,
	}
}

func (m *Manager) Create(key ChallengeKey, userID int64, generated GeneratedChallenge) {
	challenge := Challenge{
		UserID:    userID,
		ExpiresAt: time.Now().Add(m.timeout),
		Options:   generated.Options,
		correct:   generated.Correct,
	}

	m.mu.Lock()
	m.challenges[key] = challenge
	m.mu.Unlock()

	m.logger.Debug("CAPTCHA created",
		"chat_id", key.ChatID,
		"message_id", key.MessageID,
		"user_id", userID,
		"expires_at", challenge.ExpiresAt,
	)
}

func (m *Manager) Update(key ChallengeKey, generated GeneratedChallenge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.challenges[key]
	if !ok {
		return
	}
	c.Options = generated.Options
	c.correct = generated.Correct
	m.challenges[key] = c

	m.logger.Debug("CAPTCHA created", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", c.UserID, "expires_at", c.ExpiresAt)
}

func (m *Manager) Check(key ChallengeKey, userID int64, token string) CheckResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, ok := m.challenges[key]
	if !ok {
		m.logger.Debug("CAPTCHA not found", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID)
		return CheckExpired
	}

	if time.Now().After(challenge.ExpiresAt) {
		m.logger.Debug("CAPTCHA expired during check", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID)
		return CheckExpired

	}

	if challenge.UserID != userID {
		m.logger.Warn("CAPTCHA belongs to another user", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "owner_id", challenge.UserID)
		return CheckForeignUser
	}

	if challenge.correct == token {
		delete(m.challenges, key)
		m.logger.Info("CAPTCHA passed", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.Attempts)
		return CheckPassed
	}

	challenge.Attempts++

	if challenge.Attempts >= KickMaxAttempts {
		m.logger.Debug("CAPTCHA attempts limit reached", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.Attempts)
		return CheckMaxAttempts
	}

	if challenge.Attempts >= MaxAttempts {
		m.challenges[key] = challenge
		m.logger.Debug("CAPTCHA attempts limit reached", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.Attempts)
		return CheckRetry
	}

	m.challenges[key] = challenge
	m.logger.Info("CAPTCHA answer incorrect", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.Attempts)
	return CheckIncorrect
}

func (m *Manager) Delete(key ChallengeKey) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.challenges[key]; !ok {
		return
	}

	delete(m.challenges, key)

	m.logger.Debug("CAPTCHA deleted",
		"chat_id", key.ChatID,
		"message_id", key.MessageID,
	)
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
	}

	return expired
}
