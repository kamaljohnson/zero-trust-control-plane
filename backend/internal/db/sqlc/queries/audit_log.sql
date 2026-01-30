-- name: GetAuditLog :one
SELECT id, org_id, user_id, action, resource, ip, metadata, created_at
FROM audit_logs
WHERE id = $1;

-- name: ListAuditLogsByOrg :many
SELECT id, org_id, user_id, action, resource, ip, metadata, created_at
FROM audit_logs
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (id, org_id, user_id, action, resource, ip, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
