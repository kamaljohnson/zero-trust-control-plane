-- name: GetSession :one
SELECT id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, refresh_jti, refresh_token_hash, created_at
FROM sessions
WHERE id = $1;

-- name: ListSessionsByUserAndOrg :many
SELECT id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, refresh_jti, refresh_token_hash, created_at
FROM sessions
WHERE user_id = $1 AND org_id = $2 AND revoked_at IS NULL
ORDER BY created_at;

-- name: ListSessionsByOrg :many
SELECT id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, created_at
FROM sessions
WHERE org_id = $1 AND revoked_at IS NULL
  AND (sqlc.narg('user_id')::text IS NULL OR user_id = sqlc.narg('user_id'))
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: RevokeAllSessionsByUserAndOrg :exec
UPDATE sessions
SET revoked_at = $3
WHERE user_id = $1 AND org_id = $2;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, org_id, device_id, expires_at, revoked_at, last_seen_at, ip_address, refresh_jti, refresh_token_hash, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = $2
WHERE id = $1
RETURNING *;

-- name: RevokeAllSessionsByUser :exec
UPDATE sessions
SET revoked_at = $2
WHERE user_id = $1;

-- name: UpdateSessionLastSeen :one
UPDATE sessions
SET last_seen_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateSessionRefreshToken :one
UPDATE sessions
SET refresh_jti = $2, refresh_token_hash = $3
WHERE id = $1
RETURNING *;
