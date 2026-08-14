-- name: CreateCategory :one
INSERT INTO categories (name, parent_id)
VALUES ($1, $2)
RETURNING *;
