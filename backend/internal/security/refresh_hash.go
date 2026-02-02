package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashRefreshToken returns a SHA-256 hash of the refresh token string, hex-encoded.
// Used for storing and comparing refresh tokens without storing the raw token.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// RefreshTokenHashEqual performs constant-time comparison of the provided token's hash
// with the stored hash. Returns true only if they match.
func RefreshTokenHashEqual(providedToken, storedHash string) bool {
	providedHash := HashRefreshToken(providedToken)
	return subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) == 1
}
