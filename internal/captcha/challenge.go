package captcha

import (
	"time"
)

type ChallengeKey struct {
	ChatID    int64
	MessageID int64
}

type Challenge struct {
	UserID   int64
	timer    *time.Timer
	attempts int
	correct  string
}

type CheckResult int

const (
	CheckIncorrect CheckResult = iota
	CheckPassed
	CheckForeignUser
	CheckExpired
	CheckRetry
	CheckMaxAttempts
)

const (
	CountAnswer = 4

	MaxAttempts     = 2
	KickMaxAttempts = 5

	MinNumber = 4
	MaxNumber = 38
)
