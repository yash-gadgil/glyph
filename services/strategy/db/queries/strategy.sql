-- name: GetStrategiesForUser :many
SELECT * FROM strategies
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetStrategy :one
SELECT * FROM strategies
WHERE id = $1 AND user_id = $2;

-- name: CreateStrategy :one
INSERT INTO strategies (user_id, name, config)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateStrategy :one
UPDATE strategies
SET name = $3, config = $4, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteStrategy :exec
DELETE FROM strategies
WHERE id = $1 AND user_id = $2;
