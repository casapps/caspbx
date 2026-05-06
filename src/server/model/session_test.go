package model

import (
	"testing"
	"time"
)

func TestSessionIsExpired(t *testing.T) {
	now := time.Unix(100, 0)
	if !(Session{ExpiresAt: now.Add(-time.Second)}).IsExpired(now) {
		t.Fatalf("expected expired session")
	}
	if (Session{ExpiresAt: now.Add(time.Second)}).IsExpired(now) {
		t.Fatalf("expected active session")
	}
}
