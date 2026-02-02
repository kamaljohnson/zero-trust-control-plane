package security

import (
	"golang.org/x/crypto/bcrypt"
)

// Hasher hashes and verifies passwords using bcrypt. Callers must not log or
// persist plaintext passwords.
type Hasher struct {
	Cost int
}

// NewHasher returns a Hasher with the given bcrypt cost (4–31). Cost 12 is a
// reasonable default for interactive login.
func NewHasher(cost int) *Hasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	return &Hasher{Cost: cost}
}

// Hash produces a bcrypt hash of password. Uses constant-time hashing; do not
// pass empty or nil password. Returns the hash as a string suitable for storage.
func (h *Hasher) Hash(password []byte) (string, error) {
	b, err := bcrypt.GenerateFromPassword(password, h.Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare verifies password against the stored hash using constant-time
// comparison. Returns nil if they match; returns an error (including
// bcrypt.ErrMismatchedHashAndPassword) if they do not or on invalid hash.
func (h *Hasher) Compare(hash string, password []byte) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), password)
}
