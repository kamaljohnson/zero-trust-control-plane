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
	"zero-trust-control-plane/backend/internal/mfa"
	mfadomain "zero-trust-control-plane/backend/internal/mfa/domain"
	mfaintentdomain "zero-trust-control-plane/backend/internal/mfaintent/domain"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	platformsettingsdomain "zero-trust-control-plane/backend/internal/platformsettings/domain"
	"zero-trust-control-plane/backend/internal/policy/engine"
	"zero-trust-control-plane/backend/internal/security"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	sessiondomain "zero-trust-control-plane/backend/internal/session/domain"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

// Sentinel errors for auth service; handler maps them to gRPC codes.
var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidRefreshToken    = errors.New("invalid or expired refresh token")
	ErrRefreshTokenReuse      = errors.New("refresh token reuse detected; all sessions revoked")
	ErrNotOrgMember           = errors.New("user is not a member of the organization")
	ErrPhoneRequiredForMFA    = errors.New("phone number required for MFA; add in profile")
	ErrInvalidMFAChallenge    = errors.New("invalid or expired MFA challenge")
	ErrInvalidMFAIntent       = errors.New("invalid or expired MFA intent")
	ErrInvalidOTP             = errors.New("invalid OTP")
	ErrChallengeExpired       = errors.New("MFA challenge expired")
)

// AuthResult holds the outcome of Register (user_id only), Login, Refresh, or VerifyMFA (tokens + user/org).
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
	OrgID        string
}

// DevOTPStore stores plain OTP by challenge_id for dev-only retrieval (GET /dev/mfa/otp). Optional; when nil, dev OTP is not used.
type DevOTPStore interface {
	Put(ctx context.Context, challengeID, otp string, expiresAt time.Time)
}

// MFARequiredResult holds challenge_id and phone_mask when Login requires MFA before issuing a session.
type MFARequiredResult struct {
	ChallengeID string
	PhoneMask   string
}

// PhoneRequiredResult holds intent_id when Login requires MFA but the user has no phone; client must collect phone then call SubmitPhoneAndRequestMFA.
type PhoneRequiredResult struct {
	IntentID string
}

// LoginResult is the result of Login: either tokens, MFA required (challenge_id), or phone required (intent_id).
type LoginResult struct {
	Tokens        *AuthResult
	MFARequired   *MFARequiredResult
	PhoneRequired *PhoneRequiredResult
}

// UserRepo is the minimal user repository needed by the auth service.
type UserRepo interface {
	GetByID(ctx context.Context, id string) (*userdomain.User, error)
	GetByEmail(ctx context.Context, email string) (*userdomain.User, error)
	Create(ctx context.Context, u *userdomain.User) error
	SetPhoneVerified(ctx context.Context, userID, phone string) error
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
	GetByID(ctx context.Context, id string) (*devicedomain.Device, error)
	GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*devicedomain.Device, error)
	Create(ctx context.Context, d *devicedomain.Device) error
	UpdateTrustedWithExpiry(ctx context.Context, id string, trusted bool, trustedUntil *time.Time) error
}

// PlatformSettingsRepo returns platform-level device trust/MFA settings.
type PlatformSettingsRepo interface {
	GetDeviceTrustSettings(ctx context.Context, defaultTrustTTLDays int) (*platformsettingsdomain.PlatformDeviceTrustSettings, error)
}

// OrgMFASettingsRepo returns org-level MFA/device trust settings.
type OrgMFASettingsRepo interface {
	GetByOrgID(ctx context.Context, orgID string) (*orgmfasettingsdomain.OrgMFASettings, error)
}

// MFAChallengeRepo persists MFA OTP challenges.
type MFAChallengeRepo interface {
	Create(ctx context.Context, c *mfadomain.Challenge) error
	GetByID(ctx context.Context, id string) (*mfadomain.Challenge, error)
	Delete(ctx context.Context, id string) error
}

// MFAIntentRepo persists one-time MFA intents (collect phone then send OTP when user has no phone).
type MFAIntentRepo interface {
	Create(ctx context.Context, i *mfaintentdomain.Intent) error
	GetByID(ctx context.Context, id string) (*mfaintentdomain.Intent, error)
	Delete(ctx context.Context, id string) error
}

// OTPSender sends OTP via SMS (e.g. SMS Local PoC).
type OTPSender interface {
	SendOTP(phone, otp string) error
}

// MembershipRepo is the minimal membership repository needed by the auth service.
type MembershipRepo interface {
	GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error)
}

// PolicyEvaluator evaluates device-trust/MFA policies (e.g. OPA-based).
type PolicyEvaluator interface {
	EvaluateMFA(
		ctx context.Context,
		platformSettings *platformsettingsdomain.PlatformDeviceTrustSettings,
		orgSettings *orgmfasettingsdomain.OrgMFASettings,
		device *devicedomain.Device,
		user *userdomain.User,
		isNewDevice bool,
	) (engine.MFAResult, error)
}

// AuthService implements password-only register, login (with risk-based MFA), refresh, and logout.
type AuthService struct {
	userRepo             UserRepo
	identityRepo         IdentityRepo
	sessionRepo          SessionRepo
	deviceRepo           DeviceRepo
	membershipRepo       MembershipRepo
	platformSettingsRepo PlatformSettingsRepo
	orgMFASettingsRepo   OrgMFASettingsRepo
	mfaChallengeRepo     MFAChallengeRepo
	mfaIntentRepo        MFAIntentRepo
	policyEvaluator      PolicyEvaluator
	smsSender            OTPSender
	hasher               *security.Hasher
	tokens               *security.TokenProvider
	accessTTL            time.Duration
	refreshTTL           time.Duration
	defaultTrustTTLDays  int
	mfaChallengeTTL      time.Duration
	otpReturnToClient    bool
	devOTPStore          DevOTPStore
}

// NewAuthService returns an AuthService with the given dependencies.
func NewAuthService(
	userRepo UserRepo,
	identityRepo IdentityRepo,
	sessionRepo SessionRepo,
	deviceRepo DeviceRepo,
	membershipRepo MembershipRepo,
	platformSettingsRepo PlatformSettingsRepo,
	orgMFASettingsRepo OrgMFASettingsRepo,
	mfaChallengeRepo MFAChallengeRepo,
	mfaIntentRepo MFAIntentRepo,
	policyEvaluator PolicyEvaluator,
	smsSender OTPSender,
	hasher *security.Hasher,
	tokens *security.TokenProvider,
	accessTTL, refreshTTL time.Duration,
	defaultTrustTTLDays int,
	mfaChallengeTTL time.Duration,
	otpReturnToClient bool,
	devOTPStore DevOTPStore,
) *AuthService {
	if mfaChallengeTTL <= 0 {
		mfaChallengeTTL = 10 * time.Minute
	}
	return &AuthService{
		userRepo:             userRepo,
		identityRepo:         identityRepo,
		sessionRepo:          sessionRepo,
		deviceRepo:           deviceRepo,
		membershipRepo:       membershipRepo,
		platformSettingsRepo: platformSettingsRepo,
		orgMFASettingsRepo:   orgMFASettingsRepo,
		mfaChallengeRepo:     mfaChallengeRepo,
		mfaIntentRepo:        mfaIntentRepo,
		policyEvaluator:      policyEvaluator,
		smsSender:            smsSender,
		hasher:               hasher,
		tokens:               tokens,
		accessTTL:            accessTTL,
		refreshTTL:           refreshTTL,
		defaultTrustTTLDays:  defaultTrustTTLDays,
		mfaChallengeTTL:      mfaChallengeTTL,
		otpReturnToClient:    otpReturnToClient,
		devOTPStore:          devOTPStore,
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

// Login authenticates with email/password and org_id. If policy requires MFA (new/untrusted device or org/platform setting), returns MFARequired with challenge_id; otherwise creates a session and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password, orgID, deviceFingerprint string) (*LoginResult, error) {
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
	isNewDevice := dev == nil
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
	var platformSettings *platformsettingsdomain.PlatformDeviceTrustSettings
	if s.platformSettingsRepo != nil {
		platformSettings, _ = s.platformSettingsRepo.GetDeviceTrustSettings(ctx, s.defaultTrustTTLDays)
	}
	if platformSettings == nil {
		platformSettings = &platformsettingsdomain.PlatformDeviceTrustSettings{
			MFARequiredAlways:   false,
			DefaultTrustTTLDays: s.defaultTrustTTLDays,
		}
	}
	var orgSettings *orgmfasettingsdomain.OrgMFASettings
	if s.orgMFASettingsRepo != nil {
		orgSettings, _ = s.orgMFASettingsRepo.GetByOrgID(ctx, orgID)
	}
	var result engine.MFAResult
	if s.policyEvaluator != nil {
		result, _ = s.policyEvaluator.EvaluateMFA(ctx, platformSettings, orgSettings, dev, user, isNewDevice)
	} else {
		// Fallback to default behavior if no evaluator
		result = engine.MFAResult{
			MFARequired:           false,
			RegisterTrustAfterMFA: true,
			TrustTTLDays:          s.defaultTrustTTLDays,
		}
		if platformSettings != nil {
			result.TrustTTLDays = platformSettings.DefaultTrustTTLDays
		}
		if orgSettings != nil {
			result.RegisterTrustAfterMFA = orgSettings.RegisterTrustAfterMFA
			if orgSettings.TrustTTLDays > 0 {
				result.TrustTTLDays = orgSettings.TrustTTLDays
			}
		}
	}
	if result.MFARequired {
		phone := strings.TrimSpace(user.Phone)
		if phone == "" {
			// User has no phone: return intent so client can collect phone, then call SubmitPhoneAndRequestMFA.
			if s.mfaIntentRepo == nil {
				return nil, ErrPhoneRequiredForMFA
			}
			intentID := uuid.New().String()
			now := time.Now().UTC()
			expiresAt := now.Add(s.mfaChallengeTTL)
			intent := &mfaintentdomain.Intent{
				ID:        intentID,
				UserID:    user.ID,
				OrgID:     orgID,
				DeviceID:  dev.ID,
				ExpiresAt: expiresAt,
			}
			if err := s.mfaIntentRepo.Create(ctx, intent); err != nil {
				return nil, err
			}
			return &LoginResult{
				PhoneRequired: &PhoneRequiredResult{IntentID: intentID},
			}, nil
		}
		otp, err := mfa.GenerateOTP()
		if err != nil {
			return nil, err
		}
		challengeID := uuid.New().String()
		now := time.Now().UTC()
		expiresAt := now.Add(s.mfaChallengeTTL)
		challenge := &mfadomain.Challenge{
			ID:        challengeID,
			UserID:    user.ID,
			OrgID:     orgID,
			DeviceID:  dev.ID,
			Phone:     phone,
			CodeHash:  mfa.HashOTP(otp),
			ExpiresAt: expiresAt,
			CreatedAt: now,
		}
		if err := s.mfaChallengeRepo.Create(ctx, challenge); err != nil {
			return nil, err
		}
		if s.otpReturnToClient && s.devOTPStore != nil {
			s.devOTPStore.Put(ctx, challengeID, otp, expiresAt)
		} else if s.smsSender != nil {
			if err := s.smsSender.SendOTP(phone, otp); err != nil {
				_ = s.mfaChallengeRepo.Delete(ctx, challengeID)
				return nil, err
			}
		}
		phoneMask := maskPhone(phone)
		return &LoginResult{
			MFARequired: &MFARequiredResult{ChallengeID: challengeID, PhoneMask: phoneMask},
		}, nil
	}
	// MFA not required: create session without changing device trust (trust only set after MFA).
	return s.createSessionAndResult(ctx, user.ID, orgID, dev.ID, false, 0)
}

// createSessionAndResult creates a session for the given user/org/device and returns tokens. If registerTrust is true, sets device trusted with trustTTLDays.
func (s *AuthService) createSessionAndResult(ctx context.Context, userID, orgID, deviceID string, registerTrust bool, trustTTLDays int) (*LoginResult, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	refreshToken, jti, _, err := s.tokens.IssueRefresh(sessionID, userID, orgID)
	if err != nil {
		return nil, err
	}
	accessToken, _, accessExp, err := s.tokens.IssueAccess(sessionID, userID, orgID)
	if err != nil {
		return nil, err
	}
	sess := &sessiondomain.Session{
		ID:               sessionID,
		UserID:           userID,
		OrgID:            orgID,
		DeviceID:         deviceID,
		ExpiresAt:        expiresAt,
		RefreshJti:       jti,
		RefreshTokenHash: security.HashRefreshToken(refreshToken),
		CreatedAt:        time.Now().UTC(),
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}
	if registerTrust && trustTTLDays > 0 {
		trustedUntil := time.Now().UTC().AddDate(0, 0, trustTTLDays)
		_ = s.deviceRepo.UpdateTrustedWithExpiry(ctx, deviceID, true, &trustedUntil)
	}
	return &LoginResult{
		Tokens: &AuthResult{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    accessExp,
			UserID:       userID,
			OrgID:        orgID,
		},
	}, nil
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}

// SubmitPhoneAndRequestMFA consumes the intent, creates an MFA challenge for the submitted phone, sends OTP, and returns challenge_id and phone_mask.
func (s *AuthService) SubmitPhoneAndRequestMFA(ctx context.Context, intentID, phone string) (*MFARequiredResult, error) {
	intentID = strings.TrimSpace(intentID)
	phone = strings.TrimSpace(phone)
	if intentID == "" || phone == "" {
		return nil, ErrInvalidMFAIntent
	}
	if err := validatePhone(phone); err != nil {
		return nil, err
	}
	intent, err := s.mfaIntentRepo.GetByID(ctx, intentID)
	if err != nil {
		return nil, err
	}
	if intent == nil {
		return nil, ErrInvalidMFAIntent
	}
	now := time.Now().UTC()
	if !intent.ExpiresAt.After(now) {
		_ = s.mfaIntentRepo.Delete(ctx, intentID)
		return nil, ErrInvalidMFAIntent
	}
	_ = s.mfaIntentRepo.Delete(ctx, intentID)
	usr, _ := s.userRepo.GetByID(ctx, intent.UserID)
	if usr != nil && usr.PhoneVerified {
		return nil, ErrInvalidMFAIntent
	}
	otp, err := mfa.GenerateOTP()
	if err != nil {
		return nil, err
	}
	challengeID := uuid.New().String()
	expiresAt := now.Add(s.mfaChallengeTTL)
	challenge := &mfadomain.Challenge{
		ID:        challengeID,
		UserID:    intent.UserID,
		OrgID:     intent.OrgID,
		DeviceID:  intent.DeviceID,
		Phone:     phone,
		CodeHash:  mfa.HashOTP(otp),
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if err := s.mfaChallengeRepo.Create(ctx, challenge); err != nil {
		return nil, err
	}
	if s.otpReturnToClient && s.devOTPStore != nil {
		s.devOTPStore.Put(ctx, challengeID, otp, expiresAt)
	} else if s.smsSender != nil {
		if err := s.smsSender.SendOTP(phone, otp); err != nil {
			_ = s.mfaChallengeRepo.Delete(ctx, challengeID)
			return nil, err
		}
	}
	phoneMask := maskPhone(phone)
	return &MFARequiredResult{ChallengeID: challengeID, PhoneMask: phoneMask}, nil
}

// VerifyMFA verifies the OTP for the given challenge, creates a session, and optionally marks the device trusted. Returns tokens.
func (s *AuthService) VerifyMFA(ctx context.Context, challengeID, otp string) (*AuthResult, error) {
	challengeID = strings.TrimSpace(challengeID)
	otp = strings.TrimSpace(otp)
	if challengeID == "" || otp == "" {
		return nil, ErrInvalidMFAChallenge
	}
	challenge, err := s.mfaChallengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, ErrInvalidMFAChallenge
	}
	now := time.Now().UTC()
	if !challenge.ExpiresAt.After(now) {
		return nil, ErrChallengeExpired
	}
	if !mfa.OTPEqual(otp, challenge.CodeHash) {
		return nil, ErrInvalidOTP
	}
	usr, _ := s.userRepo.GetByID(ctx, challenge.UserID)
	if usr != nil && usr.Phone == "" {
		_ = s.userRepo.SetPhoneVerified(ctx, challenge.UserID, challenge.Phone)
	}
	var result engine.MFAResult
	if s.policyEvaluator != nil {
		// Get device for evaluation (usr already loaded above)
		dev, _ := s.deviceRepo.GetByID(ctx, challenge.DeviceID)
		var platformSettings *platformsettingsdomain.PlatformDeviceTrustSettings
		if s.platformSettingsRepo != nil {
			platformSettings, _ = s.platformSettingsRepo.GetDeviceTrustSettings(ctx, s.defaultTrustTTLDays)
		}
		var orgSettings *orgmfasettingsdomain.OrgMFASettings
		if s.orgMFASettingsRepo != nil {
			orgSettings, _ = s.orgMFASettingsRepo.GetByOrgID(ctx, challenge.OrgID)
		}
		result, _ = s.policyEvaluator.EvaluateMFA(ctx, platformSettings, orgSettings, dev, usr, false)
	} else {
		// Fallback to default behavior
		result = engine.MFAResult{RegisterTrustAfterMFA: true, TrustTTLDays: s.defaultTrustTTLDays}
		if s.platformSettingsRepo != nil {
			platformSettings, _ := s.platformSettingsRepo.GetDeviceTrustSettings(ctx, s.defaultTrustTTLDays)
			if platformSettings != nil {
				result.TrustTTLDays = platformSettings.DefaultTrustTTLDays
			}
		}
		if s.orgMFASettingsRepo != nil {
			orgSettings, _ := s.orgMFASettingsRepo.GetByOrgID(ctx, challenge.OrgID)
			if orgSettings != nil {
				result.RegisterTrustAfterMFA = orgSettings.RegisterTrustAfterMFA
				result.TrustTTLDays = orgSettings.TrustTTLDays
				if result.TrustTTLDays <= 0 {
					result.TrustTTLDays = s.defaultTrustTTLDays
				}
			}
		}
	}
	authResult, err := s.createSessionAndResult(ctx, challenge.UserID, challenge.OrgID, challenge.DeviceID, result.RegisterTrustAfterMFA, result.TrustTTLDays)
	if err != nil {
		return nil, err
	}
	_ = s.mfaChallengeRepo.Delete(ctx, challengeID)
	if authResult.Tokens == nil {
		return nil, ErrInvalidMFAChallenge
	}
	return authResult.Tokens, nil
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

func validatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}
	if len(phone) < 10 || len(phone) > 15 {
		return errors.New("phone must be 10 to 15 digits")
	}
	for i, r := range phone {
		if i == 0 && r == '+' {
			continue
		}
		if r < '0' || r > '9' {
			return errors.New("phone must contain only digits or a leading +")
		}
	}
	return nil
}
