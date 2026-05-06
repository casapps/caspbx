package model

import "time"

type TokenOwnerType string
type TokenScope string

const (
	TokenOwnerAdmin TokenOwnerType = "admin"
	TokenOwnerUser  TokenOwnerType = "user"
	TokenOwnerOrg   TokenOwnerType = "org"

	TokenScopeGlobal    TokenScope = "global"
	TokenScopeReadWrite TokenScope = "read-write"
	TokenScopeRead      TokenScope = "read"
)

type Token struct {
	ID          int64
	OwnerType   TokenOwnerType
	OwnerID     int64
	Name        string
	TokenHash   string
	TokenPrefix string
	Scope       TokenScope
	ExpiresAt   time.Time
	LastUsedAt  time.Time
	CreatedAt   time.Time
}

func (token Token) IsExpired(now time.Time) bool {
	return !token.ExpiresAt.IsZero() && !token.ExpiresAt.After(now)
}
