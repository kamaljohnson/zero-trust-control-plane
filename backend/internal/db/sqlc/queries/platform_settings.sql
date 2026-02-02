-- name: GetPlatformSetting :one
SELECT key, value_json
FROM platform_settings
WHERE key = $1;

-- name: SetPlatformSetting :one
INSERT INTO platform_settings (key, value_json)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET value_json = EXCLUDED.value_json
RETURNING *;

-- name: ListPlatformSettings :many
SELECT key, value_json
FROM platform_settings
ORDER BY key;
