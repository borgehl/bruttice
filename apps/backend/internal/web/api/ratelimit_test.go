package api

import (
	"testing"
	"time"
)

func TestRateLimiterPerKey(t *testing.T) {
	l := newRateLimiter(time.Hour, 2, 100)
	now := time.Now()
	if !l.Allow("a", now) || !l.Allow("a", now) {
		t.Fatal("two requests are within the limit")
	}
	if l.Allow("a", now) {
		t.Error("the third request must be refused")
	}
	if !l.Allow("b", now) {
		t.Error("another key has its own allowance")
	}
}

func TestRateLimiterForgetsAfterTheWindow(t *testing.T) {
	l := newRateLimiter(time.Hour, 1, 100)
	now := time.Now()
	l.Allow("a", now)
	if l.Allow("a", now.Add(59*time.Minute)) {
		t.Error("still inside the window")
	}
	if !l.Allow("a", now.Add(61*time.Minute)) {
		t.Error("the window has passed")
	}
	if len(l.byKey) != 1 {
		t.Errorf("old keys must be dropped: %v", l.byKey)
	}
}

func TestRateLimiterTotal(t *testing.T) {
	l := newRateLimiter(time.Hour, 5, 3)
	now := time.Now()
	for _, key := range []string{"a", "b", "c"} {
		if !l.Allow(key, now) {
			t.Fatalf("%s must be allowed", key)
		}
	}
	if l.Allow("d", now) {
		t.Error("the total is used up, a new address does not get past it")
	}
}

func TestRateLimiterRefusedRequestsAreNotCounted(t *testing.T) {
	l := newRateLimiter(time.Hour, 1, 100)
	now := time.Now()
	l.Allow("a", now)
	for i := 0; i < 10; i++ {
		l.Allow("a", now.Add(30*time.Minute))
	}
	if !l.Allow("a", now.Add(61*time.Minute)) {
		t.Error("retrying must not extend the wait")
	}
}
