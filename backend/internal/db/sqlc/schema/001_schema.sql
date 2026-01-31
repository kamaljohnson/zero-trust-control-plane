-- Enums (shared across contexts)
CREATE TYPE user_status AS ENUM ('active', 'disabled');
CREATE TYPE identity_provider AS ENUM ('local', 'oidc', 'saml');
CREATE TYPE org_status AS ENUM ('active', 'suspended');
CREATE TYPE role AS ENUM ('owner', 'admin', 'member');

-- Users (no FKs)
CREATE TABLE users (
    id         VARCHAR PRIMARY KEY,
    email      VARCHAR NOT NULL UNIQUE,
    name       VARCHAR,
    status     user_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
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
    id           VARCHAR PRIMARY KEY,
    user_id      VARCHAR NOT NULL REFERENCES users(id),
    org_id       VARCHAR NOT NULL REFERENCES organizations(id),
    fingerprint  VARCHAR NOT NULL,
    trusted      BOOLEAN NOT NULL,
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL
);

-- Sessions (ref users, organizations, devices)
CREATE TABLE sessions (
    id           VARCHAR PRIMARY KEY,
    user_id      VARCHAR NOT NULL REFERENCES users(id),
    org_id       VARCHAR NOT NULL REFERENCES organizations(id),
    device_id    VARCHAR NOT NULL REFERENCES devices(id),
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ,
    ip_address   VARCHAR,
    created_at   TIMESTAMPTZ NOT NULL
);

-- Policies (ref organizations)
CREATE TABLE policies (
    id         VARCHAR PRIMARY KEY,
    org_id     VARCHAR NOT NULL REFERENCES organizations(id),
    rules      TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

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
