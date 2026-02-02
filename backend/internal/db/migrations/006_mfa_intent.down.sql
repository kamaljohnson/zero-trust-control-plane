ALTER TABLE users DROP COLUMN IF EXISTS phone_verified;
DROP INDEX IF EXISTS idx_mfa_intents_expires_at;
DROP TABLE IF EXISTS mfa_intents;
