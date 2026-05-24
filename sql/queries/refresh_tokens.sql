-- name: CreateRefreshToken :one
INSERT INTO
    refresh_tokens (
        token,
        created_at,
        updated_at,
        expires_at,
        user_id
    )
VALUES ($1, NOW(), NOW(), $2, $3) RETURNING *;

-- name: GetRefreshToken :one
SELECT * from refresh_tokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    updated_at = NOW(),
    revoked_at = NOW()
WHERE
    token = $1;