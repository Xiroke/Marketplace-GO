-- name: ExistRefreshTokenByUser :one
SELECT EXISTS (
    SELECT 1 FROM refresh_tokens
    WHERE user_id = $1 AND token = $2
) as exists;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expired_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByRefreshToken :one
SELECT u.*
FROM users u
JOIN refresh_tokens rt ON rt.user_id = u.id
WHERE rt.token = $1
  AND rt.expired_at > NOW()
LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERe token = $1;
