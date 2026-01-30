-- name: GetIdentity :one
SELECT id, user_id, provider, provider_id, password_hash, created_at
FROM identities
WHERE id = $1;

-- name: GetIdentityByUserAndProvider :one
SELECT id, user_id, provider, provider_id, password_hash, created_at
FROM identities
WHERE user_id = $1 AND provider = $2;

-- name: GetIdentityByUserAndProviderID :one
SELECT id, user_id, provider, provider_id, password_hash, created_at
FROM identities
WHERE user_id = $1 AND provider = $2 AND provider_id = $3;

-- name: CreateIdentity :one
INSERT INTO identities (id, user_id, provider, provider_id, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateIdentityPasswordHash :one
UPDATE identities
SET password_hash = $2
WHERE id = $1
RETURNING *;
