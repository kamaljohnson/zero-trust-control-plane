-- name: GetOrgMFASettings :one
SELECT org_id, mfa_required_for_new_device, mfa_required_for_untrusted, mfa_required_always,
       register_trust_after_mfa, trust_ttl_days, created_at, updated_at
FROM org_mfa_settings
WHERE org_id = $1;

-- name: UpsertOrgMFASettings :one
INSERT INTO org_mfa_settings (org_id, mfa_required_for_new_device, mfa_required_for_untrusted,
                              mfa_required_always, register_trust_after_mfa, trust_ttl_days, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (org_id) DO UPDATE SET
    mfa_required_for_new_device = EXCLUDED.mfa_required_for_new_device,
    mfa_required_for_untrusted = EXCLUDED.mfa_required_for_untrusted,
    mfa_required_always = EXCLUDED.mfa_required_always,
    register_trust_after_mfa = EXCLUDED.register_trust_after_mfa,
    trust_ttl_days = EXCLUDED.trust_ttl_days,
    updated_at = EXCLUDED.updated_at
RETURNING *;
