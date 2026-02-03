-- Remove audit rows that reference the sentinel org, then remove the sentinel org.
DELETE FROM audit_logs WHERE org_id = '_system';
DELETE FROM organizations WHERE id = '_system';
