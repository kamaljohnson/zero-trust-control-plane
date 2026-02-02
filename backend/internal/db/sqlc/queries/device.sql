-- name: GetDevice :one
SELECT id, user_id, org_id, fingerprint, trusted, trusted_until, revoked_at, last_seen_at, created_at
FROM devices
WHERE id = $1;

-- name: GetDeviceByUserAndFingerprint :one
SELECT id, user_id, org_id, fingerprint, trusted, trusted_until, revoked_at, last_seen_at, created_at
FROM devices
WHERE user_id = $1 AND org_id = $2 AND fingerprint = $3;

-- name: ListDevicesByOrg :many
SELECT id, user_id, org_id, fingerprint, trusted, trusted_until, revoked_at, last_seen_at, created_at
FROM devices
WHERE org_id = $1
ORDER BY created_at;

-- name: CreateDevice :one
INSERT INTO devices (id, user_id, org_id, fingerprint, trusted, trusted_until, revoked_at, last_seen_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateDeviceTrusted :one
UPDATE devices
SET trusted = $2
WHERE id = $1
RETURNING *;

-- name: UpdateDeviceTrustedWithExpiry :one
UPDATE devices
SET trusted = $2, trusted_until = $3, revoked_at = NULL
WHERE id = $1
RETURNING *;

-- name: RevokeDevice :one
UPDATE devices
SET trusted = false, trusted_until = NULL, revoked_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateDeviceLastSeen :one
UPDATE devices
SET last_seen_at = $2
WHERE id = $1
RETURNING *;
