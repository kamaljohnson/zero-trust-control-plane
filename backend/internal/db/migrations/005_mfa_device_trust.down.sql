DROP INDEX IF EXISTS idx_mfa_challenges_expires_at;
DROP TABLE IF EXISTS mfa_challenges;
DROP TABLE IF EXISTS org_mfa_settings;
DROP TABLE IF EXISTS platform_settings;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE devices DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE devices DROP COLUMN IF EXISTS trusted_until;
