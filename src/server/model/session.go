package model

import "time"

type SessionKind string

const (
	SessionKindAdmin SessionKind = "admin"
	SessionKindUser  SessionKind = "user"
)

type Session struct {
	ID         string
	Kind       SessionKind
	SubjectID  int64
	TokenHash  string
	IPAddress  string
	UserAgent  string
	Location   string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
}

func (session Session) IsExpired(now time.Time) bool {
	return !session.ExpiresAt.After(now)
}
