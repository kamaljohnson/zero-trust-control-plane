-- name: GetMembership :one
SELECT id, user_id, org_id, role, created_at
FROM memberships
WHERE id = $1;

-- name: GetMembershipByUserAndOrg :one
SELECT id, user_id, org_id, role, created_at
FROM memberships
WHERE user_id = $1 AND org_id = $2;

-- name: ListMembershipsByOrg :many
SELECT id, user_id, org_id, role, created_at
FROM memberships
WHERE org_id = $1
ORDER BY created_at;

-- name: CreateMembership :one
INSERT INTO memberships (id, user_id, org_id, role, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteMembershipByUserAndOrg :exec
DELETE FROM memberships
WHERE user_id = $1 AND org_id = $2;

-- name: UpdateMembershipRole :one
UPDATE memberships
SET role = $3
WHERE user_id = $1 AND org_id = $2
RETURNING *;

-- name: CountOwnersByOrg :one
SELECT COUNT(*) FROM memberships
WHERE org_id = $1 AND role = 'owner';
