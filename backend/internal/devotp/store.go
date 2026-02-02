// Package devotp provides an in-memory store for OTP by challenge_id, used only when dev OTP mode is enabled (GET /dev/mfa/otp).
package devotp

import (
	"context"
	"sync"
	"time"
)

// Store holds plain OTP by challenge_id for dev-only retrieval. Not used in production.
type Store interface {
	// Put stores otp for challengeID until expiresAt. Used when creating an MFA challenge in dev mode.
	Put(ctx context.Context, challengeID, otp string, expiresAt time.Time)
	// Get returns the otp for challengeID if present and not expired. Returns ok false if missing or expired.
	Get(ctx context.Context, challengeID string) (otp string, ok bool)
}

type entry struct {
	otp       string
	expiresAt time.Time
}

// MemoryStore is an in-memory Store implementation.
type MemoryStore struct {
	mu   sync.RWMutex
	m    map[string]entry
	nowF func() time.Time
}

// NewMemoryStore returns a new in-memory dev OTP store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		m:    make(map[string]entry),
		nowF: time.Now().UTC,
	}
}

// Put stores otp for challengeID until expiresAt.
func (s *MemoryStore) Put(ctx context.Context, challengeID, otp string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[challengeID] = entry{otp: otp, expiresAt: expiresAt}
}

// Get returns the otp for challengeID if present and not expired.
func (s *MemoryStore) Get(ctx context.Context, challengeID string) (string, bool) {
	s.mu.RLock()
	e, ok := s.m[challengeID]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !e.expiresAt.After(s.nowF()) {
		s.mu.Lock()
		delete(s.m, challengeID)
		s.mu.Unlock()
		return "", false
	}
	return e.otp, true
}
