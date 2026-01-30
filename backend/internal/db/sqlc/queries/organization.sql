-- name: GetOrganization :one
SELECT id, name, status, created_at
FROM organizations
WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organizations (id, name, status, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, status = $3
WHERE id = $1
RETURNING *;
