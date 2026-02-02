-- Device trust: time-bound and revocable
ALTER TABLE devices ADD COLUMN trusted_until TIMESTAMPTZ;
ALTER TABLE devices ADD COLUMN revoked_at TIMESTAMPTZ;

-- User phone for MFA (PoC)
ALTER TABLE users ADD COLUMN phone VARCHAR;

-- Platform-level MFA/device trust settings (key-value)
CREATE TABLE platform_settings (
    key         VARCHAR PRIMARY KEY,
    value_json  TEXT NOT NULL
);

-- Org-level MFA/device trust settings (one row per org)
CREATE TABLE org_mfa_settings (
    org_id                       VARCHAR PRIMARY KEY REFERENCES organizations(id),
    mfa_required_for_new_device BOOLEAN NOT NULL DEFAULT true,
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
