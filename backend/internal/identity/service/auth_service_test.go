package service

import (
	"context"
	"sync"
	"testing"
	"time"

	devicedomain "zero-trust-control-plane/backend/internal/device/domain"
	identitydomain "zero-trust-control-plane/backend/internal/identity/domain"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/security"
	sessiondomain "zero-trust-control-plane/backend/internal/session/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

type memUserRepo struct {
	mu    sync.Mutex
	byID  map[string]*userdomain.User
	byEmail map[string]*userdomain.User
}

func (r *memUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byEmail[email], nil
}

func (r *memUserRepo) Create(ctx context.Context, u *userdomain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

type memIdentityRepo struct {
	mu sync.Mutex
	m  map[string]*identitydomain.Identity
}

func (r *memIdentityRepo) GetByUserAndProvider(ctx context.Context, userID string, provider identitydomain.IdentityProvider) (*identitydomain.Identity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, i := range r.m {
		if i.UserID == userID && i.Provider == provider {
			return i, nil
		}
	}
	return nil, nil
}

func (r *memIdentityRepo) Create(ctx context.Context, i *identitydomain.Identity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[i.ID] = i
	return nil
}

type memSessionRepo struct {
	mu sync.Mutex
	m  map[string]*sessiondomain.Session
}

func (r *memSessionRepo) GetByID(ctx context.Context, id string) (*sessiondomain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.m[id], nil
}

func (r *memSessionRepo) Create(ctx context.Context, s *sessiondomain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s2 := *s
	r.m[s.ID] = &s2
	return nil
}

func (r *memSessionRepo) Revoke(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.m[id]; ok {
		t := time.Now()
		s.RevokedAt = &t
	}
	return nil
}

func (r *memSessionRepo) RevokeAllSessionsByUser(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := time.Now()
	for _, s := range r.m {
		if s.UserID == userID {
			s.RevokedAt = &t
		}
	}
	return nil
}

func (r *memSessionRepo) UpdateRefreshToken(ctx context.Context, sessionID, jti, refreshTokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.m[sessionID]; ok {
		s.RefreshJti = jti
		s.RefreshTokenHash = refreshTokenHash
	}
	return nil
}

func (r *memSessionRepo) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	return nil
}

type memDeviceRepo struct {
	mu sync.Mutex
	m  map[string]*devicedomain.Device
}

func (r *memDeviceRepo) GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*devicedomain.Device, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, d := range r.m {
		if d.UserID == userID && d.OrgID == orgID && d.Fingerprint == fingerprint {
			return d, nil
		}
	}
	return nil, nil
}

func (r *memDeviceRepo) Create(ctx context.Context, d *devicedomain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d2 := *d
	r.m[d.ID] = &d2
	return nil
}

type memMembershipRepo struct {
	mu sync.Mutex
	m  map[string]*membershipdomain.Membership
}

func (r *memMembershipRepo) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.m {
		if m.UserID == userID && m.OrgID == orgID {
			return m, nil
		}
	}
	return nil, nil
}

func newTestAuthService(t *testing.T) (*AuthService, *memSessionRepo) {
	userRepo := &memUserRepo{byID: make(map[string]*userdomain.User), byEmail: make(map[string]*userdomain.User)}
	identityRepo := &memIdentityRepo{m: make(map[string]*identitydomain.Identity)}
	sessionRepo := &memSessionRepo{m: make(map[string]*sessiondomain.Session)}
	deviceRepo := &memDeviceRepo{m: make(map[string]*devicedomain.Device)}
	membershipRepo := &memMembershipRepo{m: make(map[string]*membershipdomain.Membership)}
	hasher := security.NewHasher(10)
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	svc := NewAuthService(userRepo, identityRepo, sessionRepo, deviceRepo, membershipRepo, hasher, tokens, 15*time.Minute, 24*time.Hour)
	return svc, sessionRepo
}

func TestAuthService_Register(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	res, err := svc.Register(ctx, "user@example.com", "Password123!abc", "User Name")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.UserID == "" {
		t.Fatal("expected user_id")
	}
	if res.AccessToken != "" || res.RefreshToken != "" {
		t.Fatal("Register should not return tokens")
	}

	_, err = svc.Register(ctx, "user@example.com", "Other123!abc", "")
	if err != ErrEmailAlreadyRegistered {
		t.Errorf("duplicate email: want ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestAuthService_RegisterValidation(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, "bad-email", "Password123!abc", "")
	if err == nil {
		t.Fatal("invalid email should fail")
	}
	_, err = svc.Register(ctx, "a@b.co", "Short1!abc", "")
	if err == nil {
		t.Fatal("short password should fail")
	}
	_, err = svc.Register(ctx, "a@b.co", "password123!abc", "")
	if err == nil {
		t.Fatal("password without uppercase should fail")
	}
	_, err = svc.Register(ctx, "a@b.co", "PASSWORD123!ABC", "")
	if err == nil {
		t.Fatal("password without lowercase should fail")
	}
	_, err = svc.Register(ctx, "a@b.co", "Password!!!!!abc", "")
	if err == nil {
		t.Fatal("password without number should fail")
	}
	_, err = svc.Register(ctx, "a@b.co", "Password1234abc", "")
	if err == nil {
		t.Fatal("password without symbol should fail")
	}
}

func TestAuthService_LoginRequiresMembership(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()
	_, _ = svc.Register(ctx, "user@example.com", "Password123!abc", "")

	_, err := svc.Login(ctx, "user@example.com", "Password123!abc", "org-1", "")
	if err != ErrNotOrgMember {
		t.Errorf("Login without membership: want ErrNotOrgMember, got %v", err)
	}
}

func TestAuthService_LoginAndRefreshAndLogout(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()
	reg, _ := svc.Register(ctx, "user@example.com", "Password123!abc", "")

	membershipRepo := svc.membershipRepo.(*memMembershipRepo)
	membershipRepo.mu.Lock()
	membershipRepo.m["m1"] = &membershipdomain.Membership{
		ID: "m1", UserID: reg.UserID, OrgID: "org-1", Role: membershipdomain.RoleMember,
		CreatedAt: time.Now(),
	}
	membershipRepo.mu.Unlock()

	loginRes, err := svc.Login(ctx, "user@example.com", "Password123!abc", "org-1", "")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginRes.AccessToken == "" || loginRes.RefreshToken == "" {
		t.Fatal("Login should return tokens")
	}
	if loginRes.UserID != reg.UserID || loginRes.OrgID != "org-1" {
		t.Errorf("Login user/org: got %q %q", loginRes.UserID, loginRes.OrgID)
	}

	refreshRes, err := svc.Refresh(ctx, loginRes.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if refreshRes.AccessToken == "" || refreshRes.RefreshToken == "" {
		t.Fatal("Refresh should return new tokens")
	}
	if refreshRes.UserID != reg.UserID || refreshRes.OrgID != "org-1" {
		t.Errorf("Refresh user/org: got %q %q", refreshRes.UserID, refreshRes.OrgID)
	}

	if err := svc.Logout(ctx, refreshRes.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	_, err = svc.Refresh(ctx, refreshRes.RefreshToken)
	if err != ErrInvalidRefreshToken {
		t.Errorf("Refresh after logout: want ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()
	_, _ = svc.Register(ctx, "user@example.com", "Password123!abc", "")
	membershipRepo := svc.membershipRepo.(*memMembershipRepo)
	membershipRepo.mu.Lock()
	membershipRepo.m["m1"] = &membershipdomain.Membership{
		ID: "m1", UserID: "will-replace", OrgID: "org-1", Role: membershipdomain.RoleMember,
		CreatedAt: time.Now(),
	}
	membershipRepo.mu.Unlock()
	userRepo := svc.userRepo.(*memUserRepo)
	var uid string
	for _, u := range userRepo.byID {
		uid = u.ID
		break
	}
	membershipRepo.mu.Lock()
	membershipRepo.m["m1"].UserID = uid
	membershipRepo.mu.Unlock()

	_, err := svc.Login(ctx, "user@example.com", "WrongPassword123!", "org-1", "")
	if err != ErrInvalidCredentials {
		t.Errorf("wrong password: want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_LogoutFromContext(t *testing.T) {
	svc, sessionRepo := newTestAuthService(t)
	ctx := context.Background()

	const sessionID = "sess-ctx-test"
	sessionRepo.mu.Lock()
	sessionRepo.m[sessionID] = &sessiondomain.Session{
		ID: sessionID, UserID: "u1", OrgID: "org-1", DeviceID: "d1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessionRepo.mu.Unlock()

	ctxWithSession := interceptors.WithIdentity(ctx, "u1", "org-1", sessionID)
	if err := svc.Logout(ctxWithSession, ""); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	sessionRepo.mu.Lock()
	s := sessionRepo.m[sessionID]
	sessionRepo.mu.Unlock()
	if s == nil {
		t.Fatal("session should still exist")
	}
	if s.RevokedAt == nil {
		t.Error("Logout with session in context should have revoked the session")
	}
}
