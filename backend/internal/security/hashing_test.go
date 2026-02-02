package security

import (
	"testing"
)

func TestHasher_HashAndCompare(t *testing.T) {
	h := NewHasher(10)
	password := []byte("secret123")
	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "" {
		t.Fatal("Hash returned empty")
	}
	if err := h.Compare(hash, password); err != nil {
		t.Fatalf("Compare: %v", err)
	}
}

func TestHasher_CompareWrongPassword(t *testing.T) {
	h := NewHasher(10)
	hash, _ := h.Hash([]byte("secret123"))
	if err := h.Compare(hash, []byte("wrong")); err == nil {
		t.Fatal("Compare with wrong password should fail")
	}
}

func TestHasher_Cost(t *testing.T) {
	h := NewHasher(12)
	if h.Cost != 12 {
		t.Errorf("Cost want 12, got %d", h.Cost)
	}
	h0 := NewHasher(0)
	if h0.Cost < 4 {
		t.Errorf("zero cost should be clamped to at least MinCost, got %d", h0.Cost)
	}
}
