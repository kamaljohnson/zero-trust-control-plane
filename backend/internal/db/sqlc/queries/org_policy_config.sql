-- name: GetOrgPolicyConfig :one
SELECT org_id, config_json, updated_at
FROM org_policy_config
WHERE org_id = $1;

-- name: UpsertOrgPolicyConfig :one
INSERT INTO org_policy_config (org_id, config_json, updated_at)
VALUES ($1, $2, $3)
ON CONFLICT (org_id) DO UPDATE SET
    config_json = EXCLUDED.config_json,
    updated_at = EXCLUDED.updated_at
RETURNING *;
