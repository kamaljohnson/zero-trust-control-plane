package interceptors

import (
	"context"
	"testing"
)

func TestWithIdentity_SetsAllValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "user-1", "org-1", "session-1")

	userID, ok := GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID should return true")
	}
	if userID != "user-1" {
		t.Errorf("user_id = %q, want %q", userID, "user-1")
	}

	orgID, ok := GetOrgID(ctx)
	if !ok {
		t.Fatal("GetOrgID should return true")
	}
	if orgID != "org-1" {
		t.Errorf("org_id = %q, want %q", orgID, "org-1")
	}

	sessionID, ok := GetSessionID(ctx)
	if !ok {
		t.Fatal("GetSessionID should return true")
	}
	if sessionID != "session-1" {
		t.Errorf("session_id = %q, want %q", sessionID, "session-1")
	}
}

func TestGetUserID_ReturnsValueWhenSet(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "user-1", "org-1", "session-1")

	userID, ok := GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID should return true")
	}
	if userID != "user-1" {
		t.Errorf("user_id = %q, want %q", userID, "user-1")
	}
}

func TestGetUserID_ReturnsFalseWhenNotSet(t *testing.T) {
	ctx := context.Background()

	userID, ok := GetUserID(ctx)
	if ok {
		t.Error("GetUserID should return false when not set")
	}
	if userID != "" {
		t.Errorf("user_id = %q, want empty string", userID)
	}
}

func TestGetOrgID_ReturnsValueWhenSet(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "user-1", "org-1", "session-1")

	orgID, ok := GetOrgID(ctx)
	if !ok {
		t.Fatal("GetOrgID should return true")
	}
	if orgID != "org-1" {
		t.Errorf("org_id = %q, want %q", orgID, "org-1")
	}
}

func TestGetOrgID_ReturnsFalseWhenNotSet(t *testing.T) {
	ctx := context.Background()

	orgID, ok := GetOrgID(ctx)
	if ok {
		t.Error("GetOrgID should return false when not set")
	}
	if orgID != "" {
		t.Errorf("org_id = %q, want empty string", orgID)
	}
}

func TestGetSessionID_ReturnsValueWhenSet(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "user-1", "org-1", "session-1")

	sessionID, ok := GetSessionID(ctx)
	if !ok {
		t.Fatal("GetSessionID should return true")
	}
	if sessionID != "session-1" {
		t.Errorf("session_id = %q, want %q", sessionID, "session-1")
	}
}

func TestGetSessionID_ReturnsFalseWhenNotSet(t *testing.T) {
	ctx := context.Background()

	sessionID, ok := GetSessionID(ctx)
	if ok {
		t.Error("GetSessionID should return false when not set")
	}
	if sessionID != "" {
		t.Errorf("session_id = %q, want empty string", sessionID)
	}
}

func TestContext_Isolation(t *testing.T) {
	ctx1 := context.Background()
	ctx1 = WithIdentity(ctx1, "user-1", "org-1", "session-1")

	ctx2 := context.Background()
	ctx2 = WithIdentity(ctx2, "user-2", "org-2", "session-2")

	// ctx1 should have its own values
	userID1, _ := GetUserID(ctx1)
	if userID1 != "user-1" {
		t.Errorf("ctx1 user_id = %q, want %q", userID1, "user-1")
	}

	// ctx2 should have its own values
	userID2, _ := GetUserID(ctx2)
	if userID2 != "user-2" {
		t.Errorf("ctx2 user_id = %q, want %q", userID2, "user-2")
	}
}

func TestWithIdentity_Chaining(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "user-1", "org-1", "session-1")
	ctx = WithIdentity(ctx, "user-2", "org-2", "session-2")

	// Last call should override
	userID, ok := GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID should return true")
	}
	if userID != "user-2" {
		t.Errorf("user_id = %q, want %q", userID, "user-2")
	}

	orgID, ok := GetOrgID(ctx)
	if !ok {
		t.Fatal("GetOrgID should return true")
	}
	if orgID != "org-2" {
		t.Errorf("org_id = %q, want %q", orgID, "org-2")
	}

	sessionID, ok := GetSessionID(ctx)
	if !ok {
		t.Fatal("GetSessionID should return true")
	}
	if sessionID != "session-2" {
		t.Errorf("session_id = %q, want %q", sessionID, "session-2")
	}
}

func TestWithIdentity_EmptyValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithIdentity(ctx, "", "", "")

	userID, ok := GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID should return true even for empty value")
	}
	if userID != "" {
		t.Errorf("user_id = %q, want empty string", userID)
	}

	orgID, ok := GetOrgID(ctx)
	if !ok {
		t.Fatal("GetOrgID should return true even for empty value")
	}
	if orgID != "" {
		t.Errorf("org_id = %q, want empty string", orgID)
	}

	sessionID, ok := GetSessionID(ctx)
	if !ok {
		t.Fatal("GetSessionID should return true even for empty value")
	}
	if sessionID != "" {
		t.Errorf("session_id = %q, want empty string", sessionID)
	}
}
