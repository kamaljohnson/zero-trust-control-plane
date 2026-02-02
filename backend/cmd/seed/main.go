// seed inserts development sample data for local testing. Run via ./scripts/seed.sh.
// Idempotent: skips inserts if the dev user (dev@example.com) already exists.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"zero-trust-control-plane/backend/internal/config"
	"zero-trust-control-plane/backend/internal/db"
	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/security"
)

// defaultRegoPolicy matches the default device-trust policy in internal/policy/engine/opa_evaluator.go.
const defaultRegoPolicy = `package ztcp.device_trust

default mfa_required = false
default register_trust_after_mfa = true
default trust_ttl_days = 30

mfa_required if {
	input.platform.mfa_required_always
}

mfa_required if {
	input.device.is_new
	input.org.mfa_required_for_new_device
}

mfa_required if {
	not input.device.is_effectively_trusted
	input.org.mfa_required_for_untrusted
}

register_trust_after_mfa = input.org.register_trust_after_mfa if {
	input.org.register_trust_after_mfa != null
}
register_trust_after_mfa = true if {
	not input.org.register_trust_after_mfa
}

trust_ttl_days = input.org.trust_ttl_days if {
	input.org.trust_ttl_days > 0
}
trust_ttl_days = input.platform.default_trust_ttl_days if {
	input.org.trust_ttl_days <= 0
	input.platform.default_trust_ttl_days > 0
}
`

const (
	devUserEmail     = "dev@example.com"
	devPassword      = "password123"
	devUserID        = "dev-user-001"
	devUser2ID       = "dev-user-002"
	devIdentityID    = "dev-identity-001"
	devIdentity2ID   = "dev-identity-002"
	devOrgID         = "dev-org-001"
	devMembershipID  = "dev-membership-001"
	devMembership2ID = "dev-membership-002"
	devDeviceID      = "dev-device-001"
	devPolicyID      = "dev-policy-001"
	memberEmail      = "member@example.com"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set; create a .env from .env.example or set DATABASE_URL")
	}

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer conn.Close()

	queries := gen.New(conn)
	ctx := context.Background()

	_, err = queries.GetUserByEmail(ctx, devUserEmail)
	if err == nil {
		log.Println("Seed already applied (dev@example.com exists). Skipping.")
		os.Exit(0)
	}
	if err != sql.ErrNoRows {
		log.Fatalf("seed check: %v", err)
	}

	hasher := security.NewHasher(cfg.BcryptCost)
	passwordHash, err := hasher.Hash([]byte(devPassword))
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	now := time.Now().UTC()
	trustedUntil := now.AddDate(0, 0, 30)

	if _, err := queries.CreateUser(ctx, gen.CreateUserParams{
		ID:            devUserID,
		Email:         devUserEmail,
		Name:          sql.NullString{String: "Dev User", Valid: true},
		Phone:         sql.NullString{},
		PhoneVerified: false,
		Status:        gen.UserStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		log.Fatalf("create dev user: %v", err)
	}

	if _, err := queries.CreateUser(ctx, gen.CreateUserParams{
		ID:            devUser2ID,
		Email:         memberEmail,
		Name:          sql.NullString{String: "Member User", Valid: true},
		Phone:         sql.NullString{},
		PhoneVerified: false,
		Status:        gen.UserStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		log.Fatalf("create member user: %v", err)
	}

	if _, err := queries.CreateIdentity(ctx, gen.CreateIdentityParams{
		ID:           devIdentityID,
		UserID:       devUserID,
		Provider:     gen.IdentityProviderLocal,
		ProviderID:   devUserEmail,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		CreatedAt:    now,
	}); err != nil {
		log.Fatalf("create dev identity: %v", err)
	}

	if _, err := queries.CreateIdentity(ctx, gen.CreateIdentityParams{
		ID:           devIdentity2ID,
		UserID:       devUser2ID,
		Provider:     gen.IdentityProviderLocal,
		ProviderID:   memberEmail,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		CreatedAt:    now,
	}); err != nil {
		log.Fatalf("create member identity: %v", err)
	}

	if _, err := queries.CreateOrganization(ctx, gen.CreateOrganizationParams{
		ID:        devOrgID,
		Name:      "Acme Dev",
		Status:    gen.OrgStatusActive,
		CreatedAt: now,
	}); err != nil {
		log.Fatalf("create org: %v", err)
	}

	if _, err := queries.CreateMembership(ctx, gen.CreateMembershipParams{
		ID:        devMembershipID,
		UserID:    devUserID,
		OrgID:     devOrgID,
		Role:      gen.RoleOwner,
		CreatedAt: now,
	}); err != nil {
		log.Fatalf("create dev membership: %v", err)
	}

	if _, err := queries.CreateMembership(ctx, gen.CreateMembershipParams{
		ID:        devMembership2ID,
		UserID:    devUser2ID,
		OrgID:     devOrgID,
		Role:      gen.RoleMember,
		CreatedAt: now,
	}); err != nil {
		log.Fatalf("create member membership: %v", err)
	}

	if _, err := queries.CreateDevice(ctx, gen.CreateDeviceParams{
		ID:           devDeviceID,
		UserID:       devUserID,
		OrgID:        devOrgID,
		Fingerprint:  "dev-fp-001",
		Trusted:      true,
		TrustedUntil: sql.NullTime{Time: trustedUntil, Valid: true},
		RevokedAt:    sql.NullTime{},
		LastSeenAt:   sql.NullTime{Time: now, Valid: true},
		CreatedAt:    now,
	}); err != nil {
		log.Fatalf("create device: %v", err)
	}

	if _, err := queries.UpsertOrgMFASettings(ctx, gen.UpsertOrgMFASettingsParams{
		OrgID:                   devOrgID,
		MfaRequiredForNewDevice: true,
		MfaRequiredForUntrusted: true,
		MfaRequiredAlways:       false,
		RegisterTrustAfterMfa:   true,
		TrustTtlDays:            30,
		CreatedAt:               now,
		UpdatedAt:               now,
	}); err != nil {
		log.Fatalf("upsert org mfa settings: %v", err)
	}

	if _, err := queries.SetPlatformSetting(ctx, gen.SetPlatformSettingParams{
		Key:       "default_trust_ttl_days",
		ValueJson: "30",
	}); err != nil {
		log.Fatalf("set platform setting: %v", err)
	}

	if _, err := queries.CreatePolicy(ctx, gen.CreatePolicyParams{
		ID:        devPolicyID,
		OrgID:     devOrgID,
		Rules:     defaultRegoPolicy,
		Enabled:   true,
		CreatedAt: now,
	}); err != nil {
		log.Fatalf("create policy: %v", err)
	}

	log.Println("Seed completed successfully.")
	fmt.Printf("Dev login: %s / %s\n", devUserEmail, devPassword)
	fmt.Printf("Member login: %s / %s\n", memberEmail, devPassword)
}
