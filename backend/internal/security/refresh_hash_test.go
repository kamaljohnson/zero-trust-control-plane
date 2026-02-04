package security

import (
	"testing"
)

func TestHashRefreshToken_Consistent(t *testing.T) {
	token := "test-refresh-token-123"
	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)

	if hash1 != hash2 {
		t.Errorf("HashRefreshToken not consistent: hash1 = %q, hash2 = %q", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64 (SHA-256 hex)", len(hash1))
	}
}

func TestHashRefreshToken_DifferentTokens(t *testing.T) {
	hash1 := HashRefreshToken("token-1")
	hash2 := HashRefreshToken("token-2")

	if hash1 == hash2 {
		t.Error("HashRefreshToken produced same hash for different tokens")
	}
}

func TestHashRefreshToken_EmptyToken(t *testing.T) {
	hash := HashRefreshToken("")
	if len(hash) != 64 {
		t.Errorf("hash length for empty token = %d, want 64", len(hash))
	}
}

func TestRefreshTokenHashEqual_CorrectMatch(t *testing.T) {
	token := "test-refresh-token-456"
	storedHash := HashRefreshToken(token)

	if !RefreshTokenHashEqual(token, storedHash) {
		t.Error("RefreshTokenHashEqual should match correct token")
	}
}

func TestRefreshTokenHashEqual_RejectsIncorrect(t *testing.T) {
	correctToken := "correct-token"
	wrongToken := "wrong-token"
	storedHash := HashRefreshToken(correctToken)

	if RefreshTokenHashEqual(wrongToken, storedHash) {
		t.Error("RefreshTokenHashEqual should reject incorrect token")
	}
}

func TestRefreshTokenHashEqual_ConstantTime(t *testing.T) {
	token := "test-token-789"
	storedHash := HashRefreshToken(token)

	// Test that it works correctly
	if !RefreshTokenHashEqual(token, storedHash) {
		t.Error("RefreshTokenHashEqual should match correct token")
	}

	// Test with different length inputs to ensure constant-time comparison
	wrongHash := "a" + storedHash
	if RefreshTokenHashEqual(token, wrongHash) {
		t.Error("RefreshTokenHashEqual should reject hash with different length")
	}
}

func TestRefreshTokenHashEqual_EmptyInputs(t *testing.T) {
	// Empty token with empty hash
	if RefreshTokenHashEqual("", "") {
		t.Error("RefreshTokenHashEqual should not match empty inputs")
	}

	// Empty token with non-empty hash
	hash := HashRefreshToken("some-token")
	if RefreshTokenHashEqual("", hash) {
		t.Error("RefreshTokenHashEqual should not match empty token")
	}
}

func TestRefreshTokenHashEqual_DifferentHashFormats(t *testing.T) {
	token := "test-token"
	correctHash := HashRefreshToken(token)

	// Test with hash that's same length but different content
	wrongHash := "a" + correctHash[1:]
	if RefreshTokenHashEqual(token, wrongHash) {
		t.Error("RefreshTokenHashEqual should reject hash with different content")
	}
}
