-- name: CreateMFAIntent :one
INSERT INTO mfa_intents (id, user_id, org_id, device_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMFAIntent :one
SELECT id, user_id, org_id, device_id, expires_at
FROM mfa_intents
WHERE id = $1;

-- name: DeleteMFAIntent :exec
DELETE FROM mfa_intents
WHERE id = $1;
