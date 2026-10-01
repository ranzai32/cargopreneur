package httpapi

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	if !limiter.Allow("1.2.3.4:1") || !limiter.Allow("1.2.3.4:1") {
		t.Fatal("expected the first two attempts to be allowed")
	}
	if limiter.Allow("1.2.3.4:1") {
		t.Fatal("expected the third attempt to be limited")
	}
	if !limiter.Allow("5.6.7.8:1") {
		t.Fatal("expected a different address to have its own limit")
	}
}
