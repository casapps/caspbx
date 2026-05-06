package model

import (
	"testing"
	"time"
)

func TestTokenIsExpired(t *testing.T) {
	now := time.Unix(100, 0)
	if (Token{}).IsExpired(now) {
		t.Fatalf("expected zero-expiry token to remain valid")
	}
	if !(Token{ExpiresAt: now}).IsExpired(now) {
		t.Fatalf("expected token expiring now to be expired")
	}
	if !(Token{ExpiresAt: now.Add(-time.Second)}).IsExpired(now) {
		t.Fatalf("expected past token to be expired")
	}
	if (Token{ExpiresAt: now.Add(time.Second)}).IsExpired(now) {
		t.Fatalf("expected future token to remain valid")
	}
}
