-- name: CreateTelemetry :one
INSERT INTO telemetry (org_id, user_id, device_id, session_id, event_type, source, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTelemetry :one
SELECT id, org_id, user_id, device_id, session_id, event_type, source, metadata, created_at
FROM telemetry
WHERE id = $1;

-- name: ListTelemetryByOrg :many
SELECT id, org_id, user_id, device_id, session_id, event_type, source, metadata, created_at
FROM telemetry
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
