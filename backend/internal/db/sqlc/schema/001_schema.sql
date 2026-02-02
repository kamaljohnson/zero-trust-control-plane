-- Enums (shared across contexts)
CREATE TYPE user_status AS ENUM ('active', 'disabled');
CREATE TYPE identity_provider AS ENUM ('local', 'oidc', 'saml');
CREATE TYPE org_status AS ENUM ('active', 'suspended');
CREATE TYPE role AS ENUM ('owner', 'admin', 'member');

-- Users (no FKs)
CREATE TABLE users (
    id             VARCHAR PRIMARY KEY,
    email          VARCHAR NOT NULL UNIQUE,
    name           VARCHAR,
    phone          VARCHAR,
    phone_verified BOOLEAN NOT NULL DEFAULT false,
    status         user_status NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL
);

-- Identities (ref users)
CREATE TABLE identities (
    id           VARCHAR PRIMARY KEY,
    user_id      VARCHAR NOT NULL REFERENCES users(id),
    provider     identity_provider NOT NULL,
    provider_id  VARCHAR NOT NULL,
    password_hash VARCHAR,
    created_at   TIMESTAMPTZ NOT NULL
);

-- Organizations
CREATE TABLE organizations (
    id         VARCHAR PRIMARY KEY,
    name       VARCHAR NOT NULL,
    status     org_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- Memberships (ref users, organizations)
CREATE TABLE memberships (
    id         VARCHAR PRIMARY KEY,
    user_id    VARCHAR NOT NULL REFERENCES users(id),
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    role       role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- Devices (ref users, organizations)
CREATE TABLE devices (
    id            VARCHAR PRIMARY KEY,
    user_id       VARCHAR NOT NULL REFERENCES users(id),
    org_id        VARCHAR NOT NULL REFERENCES organizations(id),
    fingerprint   VARCHAR NOT NULL,
    trusted       BOOLEAN NOT NULL,
    trusted_until TIMESTAMPTZ,
    revoked_at    TIMESTAMPTZ,
    last_seen_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL
);

-- Sessions (ref users, organizations, devices)
CREATE TABLE sessions (
    id                 VARCHAR PRIMARY KEY,
    user_id            VARCHAR NOT NULL REFERENCES users(id),
    org_id             VARCHAR NOT NULL REFERENCES organizations(id),
    device_id          VARCHAR NOT NULL REFERENCES devices(id),
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    last_seen_at       TIMESTAMPTZ,
    ip_address         VARCHAR,
    refresh_jti         VARCHAR,
    refresh_token_hash VARCHAR,
    created_at         TIMESTAMPTZ NOT NULL
);

-- Policies (ref organizations)
CREATE TABLE policies (
    id         VARCHAR PRIMARY KEY,
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    rules      TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- Platform-level settings (key-value)
CREATE TABLE platform_settings (
    key        VARCHAR PRIMARY KEY,
    value_json TEXT NOT NULL
);

-- Org-level MFA/device trust settings (one row per org)
CREATE TABLE org_mfa_settings (
    org_id                       VARCHAR PRIMARY KEY REFERENCES organizations(id),
    mfa_required_for_new_device  BOOLEAN NOT NULL DEFAULT true,
    mfa_required_for_untrusted   BOOLEAN NOT NULL DEFAULT true,
    mfa_required_always          BOOLEAN NOT NULL DEFAULT false,
    register_trust_after_mfa     BOOLEAN NOT NULL DEFAULT true,
    trust_ttl_days               INTEGER NOT NULL DEFAULT 30,
    created_at                   TIMESTAMPTZ NOT NULL,
    updated_at                   TIMESTAMPTZ NOT NULL
);

-- MFA challenges (OTP flow)
CREATE TABLE mfa_challenges (
    id         VARCHAR PRIMARY KEY,
    user_id    VARCHAR NOT NULL REFERENCES users(id),
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    device_id  VARCHAR NOT NULL REFERENCES devices(id),
    phone      VARCHAR NOT NULL,
    code_hash  VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_mfa_challenges_expires_at ON mfa_challenges(expires_at);

-- MFA intents (one-time: collect phone then send OTP when user has no phone)
CREATE TABLE mfa_intents (
    id         VARCHAR PRIMARY KEY,
    user_id    VARCHAR NOT NULL REFERENCES users(id),
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    device_id  VARCHAR NOT NULL REFERENCES devices(id),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_mfa_intents_expires_at ON mfa_intents(expires_at);

-- Audit logs (ref organizations, users)
CREATE TABLE audit_logs (
    id         VARCHAR PRIMARY KEY,
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    user_id    VARCHAR REFERENCES users(id),
    action     VARCHAR NOT NULL,
    resource   VARCHAR NOT NULL,
    ip         VARCHAR NOT NULL,
    metadata   TEXT,
    created_at TIMESTAMPTZ NOT NULL
);
