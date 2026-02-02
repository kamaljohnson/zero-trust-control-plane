package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	devicedomain "zero-trust-control-plane/backend/internal/device/domain"
	identitydomain "zero-trust-control-plane/backend/internal/identity/domain"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/security"
	sessiondomain "zero-trust-control-plane/backend/internal/session/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

// Sentinel errors for auth service; handler maps them to gRPC codes.
var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidRefreshToken     = errors.New("invalid or expired refresh token")
	ErrRefreshTokenReuse       = errors.New("refresh token reuse detected; all sessions revoked")
	ErrNotOrgMember            = errors.New("user is not a member of the organization")
)

// AuthResult holds the outcome of Register (user_id only), Login, or Refresh (tokens + user/org).
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
	OrgID        string
}

// UserRepo is the minimal user repository needed by the auth service.
type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*userdomain.User, error)
	Create(ctx context.Context, u *userdomain.User) error
}

// IdentityRepo is the minimal identity repository needed by the auth service.
type IdentityRepo interface {
	GetByUserAndProvider(ctx context.Context, userID string, provider identitydomain.IdentityProvider) (*identitydomain.Identity, error)
	Create(ctx context.Context, i *identitydomain.Identity) error
}

// SessionRepo is the minimal session repository needed by the auth service.
type SessionRepo interface {
	GetByID(ctx context.Context, id string) (*sessiondomain.Session, error)
	Create(ctx context.Context, s *sessiondomain.Session) error
	Revoke(ctx context.Context, id string) error
	RevokeAllSessionsByUser(ctx context.Context, userID string) error
	UpdateRefreshToken(ctx context.Context, sessionID, jti, refreshTokenHash string) error
	UpdateLastSeen(ctx context.Context, id string, at time.Time) error
}

// DeviceRepo is the minimal device repository needed by the auth service.
type DeviceRepo interface {
	GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*devicedomain.Device, error)
	Create(ctx context.Context, d *devicedomain.Device) error
}

// MembershipRepo is the minimal membership repository needed by the auth service.
type MembershipRepo interface {
	GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error)
}

// AuthService implements password-only register, login, refresh, and logout.
type AuthService struct {
	userRepo       UserRepo
	identityRepo   IdentityRepo
	sessionRepo   SessionRepo
	deviceRepo     DeviceRepo
	membershipRepo MembershipRepo
	hasher         *security.Hasher
	tokens         *security.TokenProvider
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// NewAuthService returns an AuthService with the given dependencies.
func NewAuthService(
	userRepo UserRepo,
	identityRepo IdentityRepo,
	sessionRepo SessionRepo,
	deviceRepo DeviceRepo,
	membershipRepo MembershipRepo,
	hasher *security.Hasher,
	tokens *security.TokenProvider,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		identityRepo:   identityRepo,
		sessionRepo:   sessionRepo,
		deviceRepo:     deviceRepo,
		membershipRepo: membershipRepo,
		hasher:         hasher,
		tokens:         tokens,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
	}
}

// Register creates a user and local identity with the given email and password.
// Returns AuthResult with UserID only (no tokens/org). Caller must Login with org_id to get tokens.
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyRegistered
	}
	userID := uuid.New().String()
	now := time.Now().UTC()
	user := &userdomain.User{
		ID:        userID,
		Email:     email,
		Name:      strings.TrimSpace(name),
		Status:    userdomain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	hashed, err := s.hasher.Hash([]byte(password))
	if err != nil {
		return nil, err
	}
	identityID := uuid.New().String()
	identity := &identitydomain.Identity{
		ID:           identityID,
		UserID:       userID,
		Provider:     identitydomain.IdentityProviderLocal,
		ProviderID:   email,
		PasswordHash: hashed,
		CreatedAt:    now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, err
	}
	return &AuthResult{UserID: userID}, nil
}

// Login authenticates with email/password and org_id, creates a session, and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password, orgID, deviceFingerprint string) (*AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	orgID = strings.TrimSpace(orgID)
	if email == "" || password == "" || orgID == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != userdomain.UserStatusActive {
		return nil, ErrInvalidCredentials
	}
	ident, err := s.identityRepo.GetByUserAndProvider(ctx, user.ID, identitydomain.IdentityProviderLocal)
	if err != nil {
		return nil, err
	}
	if ident == nil || ident.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}
	if err := s.hasher.Compare(ident.PasswordHash, []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	membership, err := s.membershipRepo.GetMembershipByUserAndOrg(ctx, user.ID, orgID)
	if err != nil {
		return nil, err
	}
	if membership == nil {
		return nil, ErrNotOrgMember
	}
	fp := strings.TrimSpace(deviceFingerprint)
	if fp == "" {
		fp = "password-login"
	}
	dev, err := s.deviceRepo.GetByUserOrgAndFingerprint(ctx, user.ID, orgID, fp)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		dev = &devicedomain.Device{
			ID:          uuid.New().String(),
			UserID:      user.ID,
			OrgID:       orgID,
			Fingerprint: fp,
			Trusted:     false,
			CreatedAt:   time.Now().UTC(),
		}
		if err := s.deviceRepo.Create(ctx, dev); err != nil {
			return nil, err
		}
	}
	sessionID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	refreshToken, jti, _, err := s.tokens.IssueRefresh(sessionID, user.ID, orgID)
	if err != nil {
		return nil, err
	}
	accessToken, _, accessExp, err := s.tokens.IssueAccess(sessionID, user.ID, orgID)
	if err != nil {
		return nil, err
	}
	sess := &sessiondomain.Session{
		ID:                sessionID,
		UserID:            user.ID,
		OrgID:             orgID,
		DeviceID:          dev.ID,
		ExpiresAt:         expiresAt,
		RefreshJti:        jti,
		RefreshTokenHash:  security.HashRefreshToken(refreshToken),
		CreatedAt:         time.Now().UTC(),
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}
	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		UserID:       user.ID,
		OrgID:        orgID,
	}, nil
}

// Refresh validates the refresh token, rotates it, and returns new tokens.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}
	sessionID, jti, userID, orgID, err := s.tokens.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}
	if sess.RefreshJti != jti {
		_ = s.sessionRepo.RevokeAllSessionsByUser(ctx, userID)
		return nil, ErrRefreshTokenReuse
	}
	if sess.RefreshTokenHash != "" && !security.RefreshTokenHashEqual(refreshToken, sess.RefreshTokenHash) {
		return nil, ErrInvalidRefreshToken
	}
	now := time.Now().UTC()
	_ = s.sessionRepo.UpdateLastSeen(ctx, sessionID, now)
	newRefresh, newJti, _, err := s.tokens.IssueRefresh(sessionID, userID, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.UpdateRefreshToken(ctx, sessionID, newJti, security.HashRefreshToken(newRefresh)); err != nil {
		return nil, err
	}
	accessToken, _, accessExp, err := s.tokens.IssueAccess(sessionID, userID, orgID)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresAt:    accessExp,
		UserID:       userID,
		OrgID:        orgID,
	}, nil
}

// Logout revokes the session identified by the refresh token or by the access token in context.
// If refreshToken is non-empty, validates it and revokes that session.
// If refreshToken is empty and the auth interceptor set session_id in context (Bearer access token), revokes that session.
// Otherwise no-op.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken != "" {
		sessionID, _, _, _, err := s.tokens.ValidateRefresh(refreshToken)
		if err != nil {
			return nil
		}
		return s.sessionRepo.Revoke(ctx, sessionID)
	}
	sessionID, ok := interceptors.GetSessionID(ctx)
	if !ok {
		return nil
	}
	return s.sessionRepo.Revoke(ctx, sessionID)
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	const simpleEmail = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	ok, _ := regexp.MatchString(simpleEmail, email)
	if !ok {
		return errors.New("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasNumber = true
		case r < '0' || (r > '9' && r < 'A') || (r > 'Z' && r < 'a') || r > 'z':
			hasSymbol = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSymbol {
		return errors.New("password must contain at least one symbol")
	}
	return nil
}
