package captcha

import (
	"log/slog"
	"sync"
	"time"
)

type Manager struct {
	timeout     time.Duration
	mu          sync.RWMutex
	challenges  map[ChallengeKey]Challenge
	logger      *slog.Logger
	expiredChan chan ExpiredChallenge
}

type ExpiredChallenge struct {
	Key    ChallengeKey
	UserID int64
}

type Option struct {
	Text  string
	Token string
}

func NewManager(timeout time.Duration, logger *slog.Logger) *Manager {
	logger.Debug("creating CAPTCHA manager")

	return &Manager{
		timeout:     timeout,
		challenges:  make(map[ChallengeKey]Challenge),
		expiredChan: make(chan ExpiredChallenge, 100),
		logger:      logger,
	}
}

func (m *Manager) Put(key ChallengeKey, userID int64, c Captcha) {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, ok := m.challenges[key]
	if !ok {
		timeoutFunc := func() {
			m.mu.Lock()
			ch, exists := m.challenges[key]
			if exists {
				delete(m.challenges, key)
				m.mu.Unlock()
				m.logger.Debug("timeoutFunc", "chat_id", key.ChatID)
				m.expiredChan <- ExpiredChallenge{Key: key, UserID: ch.UserID}
				return
			}
			m.mu.Unlock()
		}

		timer := time.AfterFunc(m.timeout, timeoutFunc)

		challenge = Challenge{
			UserID: userID,
			timer:  timer,
		}
	}
	challenge.correct = c.correctAnswer
	m.challenges[key] = challenge

	m.logger.Debug("CAPTCHA created", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", challenge.UserID)
}

func (m *Manager) Check(key ChallengeKey, userID int64, token string) CheckResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, ok := m.challenges[key]
	if !ok {
		m.logger.Debug("CAPTCHA not found", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID)
		return CheckExpired
	}

	if challenge.UserID != userID {
		m.logger.Warn("CAPTCHA belongs to another user", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "owner_id", challenge.UserID)
		return CheckForeignUser
	}

	if challenge.correct == token {
		m.logger.Info("CAPTCHA passed", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.attempts)
		return CheckPassed
	}

	challenge.attempts++

	if challenge.attempts >= KickMaxAttempts {
		m.logger.Debug("CAPTCHA attempts limit reached", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.attempts)
		return CheckMaxAttempts
	}

	if challenge.attempts >= MaxAttempts {
		m.challenges[key] = challenge
		m.logger.Debug("CAPTCHA attempts limit reached", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.attempts)
		return CheckRetry
	}

	m.challenges[key] = challenge
	m.logger.Info("CAPTCHA answer incorrect", "chat_id", key.ChatID, "message_id", key.MessageID, "user_id", userID, "attempts", challenge.attempts)
	return CheckIncorrect
}

func (m *Manager) Delete(key ChallengeKey) {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, ok := m.challenges[key]
	if ok {
		challenge.timer.Stop()

		delete(m.challenges, key)
		m.logger.Debug("CAPTCHA solved, timer stopped", "chat_id", key.ChatID)
	}
}

func (m *Manager) Expired() <-chan ExpiredChallenge {
	return m.expiredChan
}
