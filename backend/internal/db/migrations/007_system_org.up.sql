-- Sentinel organization for audit events that have no org (e.g. login_failure, logout with invalid token).
INSERT INTO organizations (id, name, status, created_at)
VALUES ('_system', 'System', 'active', now())
ON CONFLICT (id) DO NOTHING;
