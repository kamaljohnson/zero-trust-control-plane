-- MFA intent: one-time binding for "collect phone then send OTP" when user has no phone
CREATE TABLE mfa_intents (
    id         VARCHAR PRIMARY KEY,
    user_id    VARCHAR NOT NULL REFERENCES users(id),
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    device_id  VARCHAR NOT NULL REFERENCES devices(id),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_mfa_intents_expires_at ON mfa_intents(expires_at);

-- Lock phone after first MFA verification (one phone per user, immutable after verification)
ALTER TABLE users ADD COLUMN phone_verified BOOLEAN NOT NULL DEFAULT false;
