package security

import (
	"testing"
	"time"
)

func TestTokenProvider_IssueAccessAndRefresh(t *testing.T) {
	p, err := NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	sessionID, userID, orgID := "s1", "u1", "o1"

	access, accessJti, exp, err := p.IssueAccess(sessionID, userID, orgID)
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if access == "" || accessJti == "" {
		t.Fatal("access token or jti empty")
	}
	if exp.Before(time.Now()) {
		t.Fatal("expires at in the past")
	}

	refresh, jti, refreshExp, err := p.IssueRefresh(sessionID, userID, orgID)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	if refresh == "" || jti == "" {
		t.Fatal("refresh token or jti empty")
	}
	if refreshExp.Before(time.Now()) {
		t.Fatal("refresh expires at in the past")
	}

	sid, jti2, uid, oid, err := p.ValidateRefresh(refresh)
	if err != nil {
		t.Fatalf("ValidateRefresh: %v", err)
	}
	if sid != sessionID || jti2 != jti || uid != userID || oid != orgID {
		t.Errorf("ValidateRefresh: got sessionID=%q jti=%q userID=%q orgID=%q", sid, jti2, uid, oid)
	}
}

func TestTokenProvider_ValidateRefreshInvalid(t *testing.T) {
	p, err := NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	_, _, _, _, err = p.ValidateRefresh("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("ValidateRefresh invalid token: want ErrInvalidToken, got %v", err)
	}
}

func TestTokenProvider_ValidateAccess(t *testing.T) {
	p, err := NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	sessionID, userID, orgID := "s1", "u1", "o1"
	access, _, _, err := p.IssueAccess(sessionID, userID, orgID)
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	sid, uid, oid, err := p.ValidateAccess(access)
	if err != nil {
		t.Fatalf("ValidateAccess: %v", err)
	}
	if sid != sessionID || uid != userID || oid != orgID {
		t.Errorf("ValidateAccess: got sessionID=%q userID=%q orgID=%q", sid, uid, oid)
	}
}

func TestTokenProvider_ValidateAccessInvalid(t *testing.T) {
	p, err := NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	_, _, _, err = p.ValidateAccess("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("ValidateAccess invalid token: want ErrInvalidToken, got %v", err)
	}
}
