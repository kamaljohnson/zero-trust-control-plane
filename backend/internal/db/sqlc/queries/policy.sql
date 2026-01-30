-- name: GetPolicy :one
SELECT id, org_id, rules, enabled, created_at
FROM policies
WHERE id = $1;

-- name: ListPoliciesByOrg :many
SELECT id, org_id, rules, enabled, created_at
FROM policies
WHERE org_id = $1
ORDER BY created_at;

-- name: CreatePolicy :one
INSERT INTO policies (id, org_id, rules, enabled, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePolicy :one
UPDATE policies
SET rules = $2, enabled = $3
WHERE id = $1
RETURNING id, org_id, rules, enabled, created_at;
