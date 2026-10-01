package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParse(t *testing.T) {
	tokens := NewTokens("test-secret-value", time.Hour)
	id := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	raw, err := tokens.Issue(id)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tokens.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("id = %s", got)
	}
}

func TestParseRejectsTamperedAndExpired(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	tokens := NewTokens("test-secret-value", time.Hour)
	tokens.now = func() time.Time { return now }
	raw, err := tokens.Issue(uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.Parse(raw + "x"); err == nil {
		t.Fatal("expected tampered token to fail")
	}

	tokens.now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, err := tokens.Parse(raw); err == nil {
		t.Fatal("expected expired token to fail")
	}
}
