package domain

import "time"

// Identity represents a user's linked identity (local, OIDC, SAML).
type Identity struct {
	ID           string
	UserID       string
	Provider     IdentityProvider
	ProviderID   string
	PasswordHash string // empty if not local
	CreatedAt    time.Time
}

type IdentityProvider string

const (
	IdentityProviderLocal IdentityProvider = "local"
	IdentityProviderOIDC  IdentityProvider = "oidc"
	IdentityProviderSAML  IdentityProvider = "saml"
)
