CREATE TABLE org_policy_config (
    org_id      VARCHAR PRIMARY KEY REFERENCES organizations(id),
    config_json TEXT NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
