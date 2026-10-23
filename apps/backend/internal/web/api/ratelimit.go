package api

import (
	"sync"
	"time"
)

// rateLimiter allows each key a few requests per window, and all keys together a few more. The
// total keeps the limit meaningful when the key (a client address) can be made up.
type rateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	perKey   int
	total    int
	byKey    map[string][]time.Time
	allTimes []time.Time
}

func newRateLimiter(window time.Duration, perKey, total int) *rateLimiter {
	return &rateLimiter{window: window, perKey: perKey, total: total, byKey: map[string][]time.Time{}}
}

// newContactLimiter is for the contact form: 5 messages an hour from one address, 40 an hour in all.
func newContactLimiter() *rateLimiter {
	return newRateLimiter(time.Hour, 5, 40)
}

// Allow records a request at now and says if it is within the limits. A refused request is not
// recorded, so waiting is always enough.
func (l *rateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	l.allTimes = recent(l.allTimes, cutoff)
	for k, times := range l.byKey {
		if times = recent(times, cutoff); len(times) == 0 {
			delete(l.byKey, k)
		} else {
			l.byKey[k] = times
		}
	}

	if len(l.allTimes) >= l.total || len(l.byKey[key]) >= l.perKey {
		return false
	}
	l.allTimes = append(l.allTimes, now)
	l.byKey[key] = append(l.byKey[key], now)
	return true
}

// recent keeps the times after cutoff. They are in order.
func recent(times []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(times) && !times[i].After(cutoff) {
		i++
	}
	return times[i:]
}
