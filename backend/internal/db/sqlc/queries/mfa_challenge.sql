-- name: CreateMFAChallenge :one
INSERT INTO mfa_challenges (id, user_id, org_id, device_id, phone, code_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetMFAChallenge :one
SELECT id, user_id, org_id, device_id, phone, code_hash, expires_at, created_at
FROM mfa_challenges
WHERE id = $1;

-- name: DeleteMFAChallenge :exec
DELETE FROM mfa_challenges
WHERE id = $1;
