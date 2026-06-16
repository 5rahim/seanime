package handlers

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

var errTooManyRequests = errors.New("too many requests")
var errTooManyAuthenticationAttempts = errors.New("too many authentication attempts")

type rateLimitWindow struct {
	count   int
	resetAt time.Time
}

type rateLimitStore struct {
	mu      sync.Mutex
	windows map[string]*rateLimitWindow
}

func newRateLimitStore() *rateLimitStore {
	return &rateLimitStore{windows: make(map[string]*rateLimitWindow)}
}

func (s *rateLimitStore) allow(key string, limit int, window time.Duration) bool {
	if s == nil {
		return true
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for existingKey, entry := range s.windows {
		if now.After(entry.resetAt) {
			delete(s.windows, existingKey)
		}
	}

	entry, ok := s.windows[key]
	if !ok || now.After(entry.resetAt) {
		s.windows[key] = &rateLimitWindow{count: 1, resetAt: now.Add(window)}
		return true
	}

	if entry.count >= limit {
		return false
	}

	entry.count++
	return true
}

func (s *rateLimitStore) reset(key string) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.windows, key)
}

var (
	authFailureRateLimits         = newRateLimitStore()
	websocketUpgradeRateLimits    = newRateLimitStore()
	authFailureWindow             = 5 * time.Minute
	websocketUpgradeWindow        = time.Minute
	maxAuthFailuresPerWindow      = 10
	maxWebsocketAttemptsPerWindow = 40
)

func authFailureRateLimitKey(req *http.Request) string {
	return "auth:" + requestClientIP(req)
}

func websocketUpgradeRateLimitKey(req *http.Request) string {
	return "ws:" + requestClientIP(req)
}
