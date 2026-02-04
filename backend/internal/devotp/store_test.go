package devotp

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_Put(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	store.Put(ctx, "challenge-1", "123456", expiresAt)

	otp, ok := store.Get(ctx, "challenge-1")
	if !ok {
		t.Fatal("Get should return OTP after Put")
	}
	if otp != "123456" {
		t.Errorf("otp = %q, want %q", otp, "123456")
	}
}

func TestMemoryStore_Get_ReturnsOTPWhenPresent(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	store.Put(ctx, "challenge-1", "123456", expiresAt)

	otp, ok := store.Get(ctx, "challenge-1")
	if !ok {
		t.Fatal("Get should return true when OTP is present")
	}
	if otp != "123456" {
		t.Errorf("otp = %q, want %q", otp, "123456")
	}
}

func TestMemoryStore_Get_ReturnsFalseWhenMissing(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	otp, ok := store.Get(ctx, "nonexistent")
	if ok {
		t.Error("Get should return false when OTP is missing")
	}
	if otp != "" {
		t.Errorf("otp = %q, want empty string", otp)
	}
}

func TestMemoryStore_Get_ReturnsFalseWhenExpired(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(-1 * time.Minute) // Expired

	store.Put(ctx, "challenge-1", "123456", expiresAt)

	otp, ok := store.Get(ctx, "challenge-1")
	if ok {
		t.Error("Get should return false when OTP is expired")
	}
	if otp != "" {
		t.Errorf("otp = %q, want empty string", otp)
	}
}

func TestMemoryStore_Get_CleansUpExpiredEntries(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(-1 * time.Minute) // Expired

	store.Put(ctx, "challenge-1", "123456", expiresAt)

	// First Get should return false and clean up
	_, ok := store.Get(ctx, "challenge-1")
	if ok {
		t.Error("Get should return false for expired OTP")
	}

	// Second Get should also return false (entry cleaned up)
	_, ok = store.Get(ctx, "challenge-1")
	if ok {
		t.Error("Get should return false after cleanup")
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	var wg sync.WaitGroup
	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			challengeID := "challenge-" + string(rune('0'+id))
			store.Put(ctx, challengeID, "123456", expiresAt)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			challengeID := "challenge-" + string(rune('0'+id))
			store.Get(ctx, challengeID)
		}(i)
	}

	wg.Wait()
	// If there's a race condition, the test will fail with -race flag
}

func TestMemoryStore_MultipleOTPs(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	store.Put(ctx, "challenge-1", "111111", expiresAt)
	store.Put(ctx, "challenge-2", "222222", expiresAt)
	store.Put(ctx, "challenge-3", "333333", expiresAt)

	otp1, ok1 := store.Get(ctx, "challenge-1")
	otp2, ok2 := store.Get(ctx, "challenge-2")
	otp3, ok3 := store.Get(ctx, "challenge-3")

	if !ok1 || otp1 != "111111" {
		t.Errorf("challenge-1: ok=%v, otp=%q", ok1, otp1)
	}
	if !ok2 || otp2 != "222222" {
		t.Errorf("challenge-2: ok=%v, otp=%q", ok2, otp2)
	}
	if !ok3 || otp3 != "333333" {
		t.Errorf("challenge-3: ok=%v, otp=%q", ok3, otp3)
	}
}

func TestMemoryStore_ExpirationBoundary(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	
	// Test with time that's just expired (1 millisecond ago) - should be expired
	// Using millisecond instead of nanosecond to avoid timing precision issues
	expiresAt := time.Now().UTC().Add(-1 * time.Millisecond)
	store.Put(ctx, "challenge-1", "123456", expiresAt)

	// Small delay to ensure time has definitely passed
	time.Sleep(2 * time.Millisecond)

	otp, ok := store.Get(ctx, "challenge-1")
	if ok {
		t.Error("Get should return false when expiresAt is in the past")
	}
	if otp != "" {
		t.Errorf("otp = %q, want empty string", otp)
	}

	// Test with future time (should be valid)
	expiresAt = time.Now().UTC().Add(1 * time.Second)
	store.Put(ctx, "challenge-2", "654321", expiresAt)

	otp, ok = store.Get(ctx, "challenge-2")
	if !ok {
		t.Error("Get should return true when expiresAt is in the future")
	}
	if otp != "654321" {
		t.Errorf("otp = %q, want %q", otp, "654321")
	}
}
