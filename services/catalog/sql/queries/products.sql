-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, description, price, attributes, creator_id, category_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductsByCreator :many
SELECT * FROM products WHERE creator_id = $1;

-- name: GetProductsByCategory :many
SELECT p.*
FROM products p
WHERE p.category_id = $1 AND p.deleted_at IS NULL;
