package captcha

import "time"

type ChallengeKey struct {
	ChatID    int64
	MessageID int64
}

type Challenge struct {
	UserID    int64
	ExpiresAt time.Time
	Attempts  int
	Options   []Option

	correct string
}

type Option struct {
	Text  string
	Token string
}

type CheckResult int

const (
	CheckIncorrect CheckResult = iota
	CheckPassed
	CheckForeignUser
	CheckExpired
	CheckRetry
)

const MaxAttempts = 2
