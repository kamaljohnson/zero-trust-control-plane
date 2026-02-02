-- name: GetUser :one
SELECT id, email, name, phone, phone_verified, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, name, phone, phone_verified, status, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, name, phone, phone_verified, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $2, name = $3, phone = $4, phone_verified = $5, status = $6, updated_at = $7
WHERE id = $1
RETURNING *;

-- name: SetPhoneVerified :one
UPDATE users
SET phone = $2, phone_verified = true, updated_at = $3
WHERE id = $1 AND (phone IS NULL OR phone = '') AND phone_verified = false
RETURNING id;
