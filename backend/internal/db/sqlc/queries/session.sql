-- name: GetSession :one
SELECT id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, created_at
FROM sessions
WHERE id = $1;

-- name: ListSessionsByUserAndOrg :many
SELECT id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, created_at
FROM sessions
WHERE user_id = $1 AND org_id = $2 AND revoked_at IS NULL
ORDER BY created_at;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateSessionLastSeen :one
UPDATE sessions
SET last_seen_at = $2
WHERE id = $1
RETURNING *;
