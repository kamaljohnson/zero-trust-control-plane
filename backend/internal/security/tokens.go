package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a token is malformed or invalid.
	ErrInvalidToken = errors.New("invalid token")
)

// AccessClaims holds JWT claims for the access token.
type AccessClaims struct {
	jwt.RegisteredClaims
	OrgID     string `json:"org_id"`
	SessionID string `json:"session_id"`
}

// RefreshClaims holds JWT claims for the refresh token (includes jti for rotation).
type RefreshClaims struct {
	jwt.RegisteredClaims
	SessionID string `json:"session_id"`
	OrgID     string `json:"org_id"`
}

// TokenProvider issues and validates JWT access and refresh tokens using RS256 or ES256 (private/public key).
type TokenProvider struct {
	privateKey crypto.Signer
	publicKey  crypto.PublicKey
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenProvider returns a TokenProvider that signs with the given private key (RS256 or ES256).
// issuer and audience are set on claims and validated on refresh.
func NewTokenProvider(privateKey crypto.Signer, publicKey crypto.PublicKey, issuer, audience string, accessTTL, refreshTTL time.Duration) *TokenProvider {
	return &TokenProvider{
		privateKey:  privateKey,
		publicKey:   publicKey,
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// IssueAccess issues a short-lived access JWT for the given session, user, and org.
// Returns the token string, its jti, and expiration time.
func (p *TokenProvider) IssueAccess(sessionID, userID, orgID string) (token string, jti string, expiresAt time.Time, err error) {
	jti, err = generateJTI()
	if err != nil {
		return "", "", time.Time{}, err
	}
	now := time.Now().UTC()
	expiresAt = now.Add(p.accessTTL)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			Issuer:    p.issuer,
			Audience:  jwt.ClaimStrings{p.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		OrgID:     orgID,
		SessionID: sessionID,
	}
	token, err = p.sign(claims)
	return token, jti, expiresAt, err
}

// IssueRefresh issues a long-lived refresh JWT and returns the token, its jti
// (for rotation binding), and expiration time. Caller should store jti on the session.
func (p *TokenProvider) IssueRefresh(sessionID, userID, orgID string) (token, jti string, expiresAt time.Time, err error) {
	jti, err = generateJTI()
	if err != nil {
		return "", "", time.Time{}, err
	}
	now := time.Now().UTC()
	expiresAt = now.Add(p.refreshTTL)
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			Issuer:    p.issuer,
			Audience:  jwt.ClaimStrings{p.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		SessionID: sessionID,
		OrgID:     orgID,
	}
	token, err = p.sign(claims)
	return token, jti, expiresAt, err
}

func (p *TokenProvider) sign(claims jwt.Claims) (string, error) {
	var method jwt.SigningMethod
	switch p.privateKey.Public().(type) {
	case *rsa.PublicKey:
		method = jwt.SigningMethodRS256
	case *ecdsa.PublicKey:
		method = jwt.SigningMethodES256
	default:
		return "", ErrInvalidToken
	}
	t := jwt.NewWithClaims(method, claims)
	return t.SignedString(p.privateKey)
}

// ValidateRefresh parses and validates the refresh token (signature, exp, iss, aud).
// Returns sessionID, jti, userID, orgID, or error.
func (p *TokenProvider) ValidateRefresh(tokenString string) (sessionID, jti, userID, orgID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return p.publicKey, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
			return p.publicKey, nil
		}
		return nil, ErrInvalidToken
	})
	if err != nil {
		return "", "", "", "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return "", "", "", "", ErrInvalidToken
	}
	if claims.Issuer != p.issuer {
		return "", "", "", "", ErrInvalidToken
	}
	audOk := false
	for _, a := range claims.Audience {
		if a == p.audience {
			audOk = true
			break
		}
	}
	if !audOk {
		return "", "", "", "", ErrInvalidToken
	}
	return claims.SessionID, claims.ID, claims.Subject, claims.OrgID, nil
}

// ValidateAccess parses and validates the access token (signature, exp, iss, aud).
// Returns sessionID, userID, orgID, or error.
func (p *TokenProvider) ValidateAccess(tokenString string) (sessionID, userID, orgID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return p.publicKey, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
			return p.publicKey, nil
		}
		return nil, ErrInvalidToken
	})
	if err != nil {
		return "", "", "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return "", "", "", ErrInvalidToken
	}
	if claims.Issuer != p.issuer {
		return "", "", "", ErrInvalidToken
	}
	audOk := false
	for _, a := range claims.Audience {
		if a == p.audience {
			audOk = true
			break
		}
	}
	if !audOk {
		return "", "", "", ErrInvalidToken
	}
	return claims.SessionID, claims.Subject, claims.OrgID, nil
}

func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
