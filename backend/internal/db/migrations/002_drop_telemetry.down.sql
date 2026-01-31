-- Telemetry (ref org; optional user, device, session)
CREATE TABLE telemetry (
    id          BIGSERIAL PRIMARY KEY,
    org_id      VARCHAR NOT NULL,
    user_id     VARCHAR,
    device_id   VARCHAR,
    session_id  VARCHAR,
    event_type  VARCHAR NOT NULL,
    source      VARCHAR NOT NULL,
    metadata    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);
